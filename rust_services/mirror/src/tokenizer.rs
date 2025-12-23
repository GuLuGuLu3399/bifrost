// rust_services/mirror/src/tokenizer.rs
//
// 基于 jieba-rs 的中文分词器实现，兼容 tantivy 0.25+

use std::sync::Arc;

use jieba_rs::Jieba;
use tantivy::tokenizer::{Token, TokenStream, Tokenizer};

/// 分词器名称常量，用于在 tantivy 中注册
pub const JIEBA_TOKENIZER: &str = "jieba";

/// 基于 jieba-rs 的中文分词器
#[derive(Clone)]
pub struct JiebaTokenizer {
    jieba: Arc<Jieba>,
}

impl Default for JiebaTokenizer {
    fn default() -> Self {
        Self {
            jieba: Arc::new(Jieba::new()),
        }
    }
}

impl Tokenizer for JiebaTokenizer {
    type TokenStream<'a> = JiebaTokenStream;

    fn token_stream<'a>(&'a mut self, text: &'a str) -> Self::TokenStream<'a> {
        let words: Vec<(String, usize, usize)> = self
            .jieba
            .tokenize(text, jieba_rs::TokenizeMode::Search, false)
            .into_iter()
            .map(|t| (t.word.to_lowercase(), t.start, t.end))
            .collect();

        JiebaTokenStream {
            tokens: words,
            index: 0,
            token: Token::default(),
        }
    }
}

/// Jieba 分词结果的 Token 流
pub struct JiebaTokenStream {
    tokens: Vec<(String, usize, usize)>,
    index: usize,
    token: Token,
}

impl TokenStream for JiebaTokenStream {
    fn advance(&mut self) -> bool {
        if self.index < self.tokens.len() {
            let (ref word, start, end) = self.tokens[self.index];
            self.token.text.clear();
            self.token.text.push_str(word);
            self.token.offset_from = start;
            self.token.offset_to = end;
            self.token.position = self.index;
            self.index += 1;
            true
        } else {
            false
        }
    }

    fn token(&self) -> &Token {
        &self.token
    }

    fn token_mut(&mut self) -> &mut Token {
        &mut self.token
    }
}
