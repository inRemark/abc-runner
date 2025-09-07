# Document Translation and Maintenance Guide

[English](document-translation-guide.md) | [中文](document-translation-guide.zh.md)

This guide provides guidelines for maintaining multilingual documentation in the redis-runner project.

## 1. File Naming Conventions

### 1.1 English Documents
- Use the `.md` extension
- Examples: `README.md`, `redis.md`, `http.md`

### 1.2 Chinese Documents
- Use the `.zh.md` extension
- Examples: `README.zh.md`, `redis.zh.md`, `http.zh.md`

## 2. Content Standards

### 2.1 Language Consistency
- English documents must contain entirely English content
- Chinese documents must contain entirely Chinese content
- Technical terms should be consistent between both language versions

### 2.2 Technical Term Handling
- Technical content such as command-line arguments, configuration items, and code examples should remain unchanged during translation
- Proper nouns (such as Redis, Kafka, HTTP, etc.) should use the original English words in Chinese documents as well

### 2.3 Format Standards
- Both language versions of documents should maintain the same structure and format
- Markdown elements such as code blocks, lists, and headings should be consistent

## 3. Maintenance Process

### 3.1 Adding New Documents
1. First create the English version of the document
2. Then create the corresponding Chinese version
3. Ensure that both versions correspond accurately

### 3.2 Updating Documents
1. When modifying documents, synchronously update the corresponding language versions
2. Use `make validate-docs` to check language consistency before committing
3. Ensure links and navigation are correct in both versions

### 3.3 Language Switching Links
Each document should include language switching links at the top:
```markdown
[English](filename.md) | [中文](filename.zh.md)
```

## 4. Quality Assurance

### 4.1 Automated Checking
- Use the `make validate-docs` command to check document language consistency
- Document validation will automatically run in the CI/CD pipeline

### 4.2 Manual Review
- Regularly review documents to ensure quality
- Invite native speakers to participate in translation quality checks

## 5. Best Practices

### 5.1 Translation Recommendations
- Maintain consistency of technical terms
- Pay attention to differences in expression habits between Chinese and English
- Ensure example code is consistent between both versions

### 5.2 Document Structure
- Maintain consistent structure between Chinese and English documents
- Chapter titles should be accurately translated
- Important content should not be omitted during translation

## 6. Common Issues

### 6.1 Language Mixing
- Ensure English documents do not contain Chinese content
- Ensure Chinese documents do not contain large sections of English content (except for technical terms)

### 6.2 Broken Links
- Check the validity of all links when updating documents
- Ensure language switching links point to the correct files

By following these guidelines, we can ensure that the multilingual documentation of the redis-runner project maintains high quality and consistency.