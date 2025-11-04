# 📝 Complaints Documentation

This directory contains automatically generated documentation of all complaints filed through the `file_complaint` tool.

## 🗂️ File Structure

Complaints are automatically exported to this directory as human-readable markdown files with the following naming convention:

```
YYYY-MM-DD_HH-MM-SESSION_NAME.md
```

- **Example**: `2025-11-04_03-10-39-simple-session.md`
- **No Session**: `2025-11-04_03-10-39-no-session.md`

## 📋 What's Included

Each complaint markdown file contains:

- **Metadata**: Creation time, session, severity, project, status
- **Task Description**: What the AI was asked to do
- **Context Information**: Available context and information
- **Missing Information**: What information was missing or unclear
- **What Confused Me**: Specific points of confusion
- **Future Wishes**: Suggestions for improvement
- **Resolution Status**: Whether and how the complaint was resolved

## ⚙️ Configuration

Documentation export can be configured via:

### Configuration File (config.yaml):
```yaml
storage:
  docs_enabled: true
  docs_dir: "docs/complaints"
  docs_format: "markdown"  # markdown, html, text
  retention_days: 0       # 0 = infinite retention
```

### Command Line Flags:
```bash
--storage.docs-enabled=true
--storage.docs-dir=docs/complaints
--storage.docs-format=markdown
--storage.retention-days=0
```

### Environment Variables:
```bash
COMPLAINTS_MCP_STORAGE_DOCS_ENABLED=true
COMPLAINTS_MCP_STORAGE_DOCS_DIR=docs/complaints
COMPLAINTS_MCP_STORAGE_DOCS_FORMAT=markdown
COMPLAINTS_MCP_STORAGE_RETENTION_DAYS=0
```

## 🔄 Export Formats

### Markdown (Default)
- Human-readable and version control friendly
- Easy to edit and contribute to
- Best for documentation and collaboration

### HTML
- Web-friendly format
- Rich presentation for browsers
- Good for internal dashboards

### Text
- Simple plain text format
- Maximum compatibility
- Good for logs and parsing

## ♾️ Retention Policy

- **Default**: Infinite retention (retention_days = 0)
- **Purpose**: Complete documentation history for continuous improvement
- **Backup**: Files are version controlled and backed up
- **Cleanup**: Only manual cleanup should remove documented complaints

## 📊 Usage

1. **Automatic**: Complaints are automatically exported when created
2. **Search**: Use file system search or grep to find specific issues
3. **Version Control**: All changes are tracked in git history
4. **Collaboration**: Edit markdown files to add additional context

## 🎯 Purpose

This documentation system serves to:

- **Track**: Record all user complaints and issues
- **Improve**: Identify patterns and areas for enhancement
- **Learn**: Build knowledge base of common issues
- **Document**: Create permanent record for future reference
- **Collaborate**: Enable team contributions to solutions

---

*This directory is automatically generated and maintained by the complaints-mcp system.*