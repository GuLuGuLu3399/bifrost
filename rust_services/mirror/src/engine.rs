// rust_services/mirror/src/engine.rs

use anyhow::Result;
use std::path::Path;
use std::sync::{Arc, RwLock};

use tantivy::collector::{Count, TopDocs};
use tantivy::query::{BooleanQuery, Occur, QueryParser, TermQuery};
use tantivy::schema::*;
use tantivy::{Index, IndexReader, IndexWriter, ReloadPolicy, TantivyDocument, Term};

use crate::tokenizer::{JiebaTokenizer, JIEBA_TOKENIZER};

// 定义字段名称常量
const FIELD_ID: &str = "id";
const FIELD_TITLE: &str = "title";
const FIELD_SLUG: &str = "slug";
const FIELD_BODY: &str = "body";
const FIELD_STATUS: &str = "status"; // 1=Draft, 2=Published
const FIELD_PUBLISHED_AT: &str = "published_at";

// ============ 搜索引擎 ============

#[derive(Clone)]
pub struct SearchEngine {
    index: Index,
    reader: IndexReader,
    // 使用 RwLock 保护 IndexWriter (单线程写入)
    writer: Arc<RwLock<IndexWriter>>,
    // 缓存字段句柄
    fields: Arc<SchemaFields>,
}

struct SchemaFields {
    id: Field,
    title: Field,
    slug: Field,
    body: Field,
    status: Field,
    published_at: Field,
}

impl SearchEngine {
    /// 初始化搜索引擎
    pub fn new(index_path: Option<&str>) -> Result<Self> {
        // 1. 定义 Schema
        let mut schema_builder = Schema::builder();

        // ID: 用于回查 Beacon
        let id = schema_builder.add_i64_field(FIELD_ID, STORED | FAST | INDEXED);

        // 配置中文分词 (Jieba)
        let text_options = TextOptions::default()
            .set_indexing_options(
                TextFieldIndexing::default()
                    .set_tokenizer(JIEBA_TOKENIZER) // 使用自定义中文分词器
                    .set_index_option(IndexRecordOption::WithFreqsAndPositions),
            )
            .set_stored();

        let title = schema_builder.add_text_field(FIELD_TITLE, text_options.clone());
        let body = schema_builder.add_text_field(FIELD_BODY, text_options);

        // Slug: 不分词，完全匹配
        let slug = schema_builder.add_text_field(FIELD_SLUG, TextOptions::default().set_stored());

        // Status & PublishedAt: 用于过滤和排序
        let status = schema_builder.add_i64_field(FIELD_STATUS, FAST | INDEXED);
        let published_at = schema_builder.add_i64_field(FIELD_PUBLISHED_AT, FAST | STORED);

        let schema = schema_builder.build();

        // 2. 打开或创建 Index
        let index = match index_path {
            Some(path) => {
                let p = Path::new(path);
                std::fs::create_dir_all(p)?;
                Index::open_or_create(tantivy::directory::MmapDirectory::open(p)?, schema.clone())?
            }
            None => Index::create_in_ram(schema.clone()),
        };

        // 注册中文分词器
        index
            .tokenizers()
            .register(JIEBA_TOKENIZER, JiebaTokenizer::default());

        // 4. 创建 Reader
        // 使用 Manual 策略，配合写入时的显式 reload，确保数据一致性
        let reader = index
            .reader_builder()
            .reload_policy(ReloadPolicy::Manual)
            .try_into()?;

        // 5. 创建 Writer (分配 50MB buffer)
        let writer = index.writer(50_000_000)?;

        Ok(Self {
            index,
            reader,
            writer: Arc::new(RwLock::new(writer)),
            fields: Arc::new(SchemaFields {
                id,
                title,
                slug,
                body,
                status,
                published_at,
            }),
        })
    }

    /// 索引文档 (Upsert: Delete + Add)
    /// [优化] 增加 published_at 参数，确保索引时间与文章实际发布时间一致
    pub fn index_doc(
        &self,
        id: i64,
        title: &str,
        body: &str,
        slug: &str,
        status: i64,
        published_at: i64,
    ) -> Result<()> {
        // 使用代码块限制 writer 锁的作用域
        {
            let mut writer = self
                .writer
                .write()
                .map_err(|_| anyhow::anyhow!("Lock poisoned"))?;

            // 1. 删除旧文档
            let id_term = Term::from_field_i64(self.fields.id, id);
            writer.delete_term(id_term);

            // 2. 添加新文档
            let mut doc = TantivyDocument::default();
            doc.add_i64(self.fields.id, id);
            doc.add_text(self.fields.title, title);
            doc.add_text(self.fields.body, body);
            doc.add_text(self.fields.slug, slug);
            doc.add_i64(self.fields.status, status);
            doc.add_i64(self.fields.published_at, published_at);

            writer.add_document(doc)?;
            writer.commit()?;
        } // writer 锁在此处释放

        // [关键修复] 显式重载 Reader，使写入立即对搜索可见
        self.reader.reload()?;

        Ok(())
    }

    /// 删除文档
    pub fn delete_doc(&self, id: i64) -> Result<()> {
        // 使用代码块限制 writer 锁的作用域
        {
            let mut writer = self
                .writer
                .write()
                .map_err(|_| anyhow::anyhow!("Lock poisoned"))?;

            let id_term = Term::from_field_i64(self.fields.id, id);
            writer.delete_term(id_term);
            writer.commit()?;
        } // writer 锁在此处释放

        // [关键修复] 删除后也需要重载
        self.reader.reload()?;

        Ok(())
    }

    /// 执行搜索
    pub fn search(&self, query_str: &str, limit: usize) -> Result<(Vec<(i64, f32)>, usize)> {
        let searcher = self.reader.searcher();

        // 1. 构建查询解析器
        let query_parser =
            QueryParser::for_index(&self.index, vec![self.fields.title, self.fields.body]);

        // 2. 解析 Query
        let query = query_parser.parse_query(query_str)?;

        // 3. 构造 Boolean Query: (UserQuery) AND (Status=Published)
        // 假设 Published 状态码为 2
        let status_term = Term::from_field_i64(self.fields.status, 2);
        let status_query = TermQuery::new(status_term, IndexRecordOption::Basic);

        let final_query = BooleanQuery::new(vec![
            (Occur::Must, query),
            (Occur::Must, Box::new(status_query)),
        ]);

        // 4. 执行检索
        let top_docs = searcher.search(&final_query, &TopDocs::with_limit(limit))?;
        let count = searcher.search(&final_query, &Count)?;

        // 5. 提取结果
        let mut results = Vec::with_capacity(top_docs.len());
        for (score, doc_address) in top_docs {
            let retrieved_doc: TantivyDocument = searcher.doc(doc_address)?;
            // 提取 ID
            if let Some(id_val) = retrieved_doc.get_first(self.fields.id) {
                if let Some(id_num) = id_val.as_i64() {
                    results.push((id_num, score));
                }
            }
        }

        Ok((results, count))
    }

    /// 检查文档是否存在于索引中 (用于调试)
    pub fn doc_exists(&self, id: i64) -> Result<bool> {
        let searcher = self.reader.searcher();

        let id_term = Term::from_field_i64(self.fields.id, id);
        let id_query = TermQuery::new(id_term, IndexRecordOption::Basic);

        let count = searcher.search(&id_query, &Count)?;

        Ok(count > 0)
    }
}
