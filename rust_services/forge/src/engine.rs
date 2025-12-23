// rust_services/forge/src/engine.rs

use anyhow::Result;
use ammonia::Builder;
use pulldown_cmark::{html, Event, HeadingLevel, Options, Parser, Tag, TagEnd};
use serde::{Deserialize, Serialize};

#[derive(Debug, Serialize, Deserialize)]
pub struct TocEntry {
    pub level: u8,
    pub title: String,
    pub anchor: String,
}

pub struct MarkdownEngine {
    cleaner: Builder<'static>,
}

impl Default for MarkdownEngine {
    fn default() -> Self {
        let mut cleaner = Builder::default();
        cleaner.add_generic_attributes(&["id", "class", "style"]);
        Self { cleaner }
    }
}

impl MarkdownEngine {
    pub fn new() -> Self {
        Self::default()
    }

    pub fn render(&self, markdown: &str) -> Result<(String, String, String)> {
        let mut options = Options::empty();
        options.insert(Options::ENABLE_TABLES);
        options.insert(Options::ENABLE_FOOTNOTES);
        options.insert(Options::ENABLE_STRIKETHROUGH);
        options.insert(Options::ENABLE_TASKLISTS);

        let parser = Parser::new_ext(markdown, options);

        let mut toc = Vec::new();
        let mut html_output = String::new();
        let mut summary = String::new();

        let mut in_heading = false;
        let mut current_header_level: u8 = 0;
        let mut current_header_text = String::new();

        let mapper = parser.inspect(|event| {
            match &event {
                Event::Start(Tag::Heading { level, id: _, classes: _, attrs: _ }) => {
                    in_heading = true;
                    current_header_level = match level {
                        HeadingLevel::H1 => 1,
                        HeadingLevel::H2 => 2,
                        HeadingLevel::H3 => 3,
                        HeadingLevel::H4 => 4,
                        HeadingLevel::H5 => 5,
                        HeadingLevel::H6 => 6,
                    };
                    current_header_text.clear();
                }
                Event::End(TagEnd::Heading(_)) => {
                    in_heading = false;
                    let anchor = slug::slugify(&current_header_text);

                    toc.push(TocEntry {
                        level: current_header_level,
                        title: current_header_text.clone(),
                        anchor: anchor.clone(),
                    });
                }
                // 摘要收集
                Event::Text(text) => {
                    // 收集纯文本作为摘要，限制长度为 200 字符
                    // TODO 后续考虑换成AI生成摘要或者摘要提取算法
                    if in_heading {
                        current_header_text.push_str(text);
                    } else if summary.chars().count() < 200 {
                        summary.push_str(text);
                    }
                }
                _ => {}
            }
        });

        html::push_html(&mut html_output, mapper);

        let safe_html = self.cleaner.clean(&html_output).to_string();
        let toc_json = serde_json::to_string(&toc).unwrap_or_else(|_| "[]".into());

        Ok((safe_html, toc_json, summary))
    }
}
