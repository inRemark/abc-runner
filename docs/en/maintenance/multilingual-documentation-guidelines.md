# Multilingual Documentation Guidelines

[English](multilingual-documentation-guidelines.md) | [中文](multilingual-documentation-guidelines.zh.md)

## Overview

This document provides guidelines for maintaining multilingual documentation in the redis-runner project. The goal is to ensure consistency and quality across all language versions of our documentation.

## Directory Structure

The documentation follows a clear directory structure to organize content by language:

```
docs/
├── en/                    # English documentation
│   ├── architecture/
│   ├── deployment/
│   ├── developer-guide/
│   ├── getting-started/
│   └── user-guide/
├── zh/                    # Chinese documentation
│   ├── architecture/
│   ├── deployment/
│   ├── developer-guide/
│   ├── getting-started/
│   └── user-guide/
└── maintenance/           # Maintenance documentation (bilingual)
```

## Language Tags

Each document should include language tags at the top to allow easy switching between languages:

```markdown
[English](document.md) | [中文](../zh/path/document.md)
```

For documents in the `maintenance` directory, both languages should be in the same directory:

```markdown
[English](document.md) | [中文](document.zh.md)
```

## Translation Workflow

### 1. New English Documents

When creating new English documents:

1. Place the document in the appropriate subdirectory under `docs/en/`
2. Create a corresponding Chinese translation in the same path under `docs/zh/`
3. Add language tags to both documents
4. Update relevant README files and navigation links

### 2. Updates to Existing Documents

When updating existing documents:

1. Make changes to both language versions
2. Ensure technical accuracy is maintained in both versions
3. Keep formatting and structure consistent between versions
4. Update any affected links or references

### 3. Adding New Languages

To add support for a new language:

1. Create a new directory under `docs/` with the appropriate language code (e.g., `docs/fr/` for French)
2. Translate existing documents to the new language
3. Update language tags in all documents to include the new language
4. Update README files to include links to the new language versions

## Quality Assurance

### Consistency Checks

1. **Technical Accuracy**: Ensure technical content is accurate in all language versions
2. **Terminology Consistency**: Use consistent terminology across all documents
3. **Formatting Consistency**: Maintain consistent formatting and structure
4. **Link Validity**: Verify that all links work correctly

### Review Process

1. **Self-Review**: Authors should review their own translations
2. **Peer Review**: Have another team member review translations
3. **Native Speaker Review**: For non-English documents, have a native speaker review for language accuracy
4. **Technical Review**: Have a technical expert verify technical content accuracy

## Best Practices

### Writing Guidelines

1. **Clear and Concise**: Use clear, concise language
2. **Active Voice**: Prefer active voice over passive voice
3. **Consistent Terminology**: Use consistent technical terms
4. **Cultural Sensitivity**: Be aware of cultural differences in examples and references

### Translation Guidelines

1. **Accuracy**: Prioritize accuracy over literal translation
2. **Natural Language**: Use natural language for the target audience
3. **Technical Terms**: Keep technical terms consistent with industry standards
4. **Cultural Adaptation**: Adapt examples and references for cultural relevance

### Code and Examples

1. **Code Comments**: Translate code comments when appropriate
2. **Output Messages**: Translate output messages and error messages
3. **File Names**: Keep file names in English for consistency
4. **Command Examples**: Keep command examples in English, but translate explanations

## Tools and Resources

### Translation Tools

1. **Machine Translation**: Use machine translation as a starting point, but always review and edit
2. **Translation Memory**: Maintain a translation memory for consistent terminology
3. **Glossary**: Maintain a glossary of technical terms and their translations

### Review Tools

1. **Spell Checkers**: Use spell checkers for all languages
2. **Grammar Checkers**: Use grammar checkers when available
3. **Link Checkers**: Regularly check links for validity

## Maintenance Schedule

### Regular Reviews

1. **Monthly**: Quick consistency check of recent changes
2. **Quarterly**: Comprehensive review of all documentation
3. **Annually**: Full audit of all documentation for accuracy and relevance

### Update Triggers

1. **Feature Releases**: Update documentation when new features are added
2. **Bug Fixes**: Update documentation when bugs are fixed
3. **User Feedback**: Update documentation based on user feedback
4. **Technology Changes**: Update documentation when underlying technologies change

## Roles and Responsibilities

### Documentation Maintainers

1. **Overall Coordination**: Coordinate documentation efforts across languages
2. **Quality Assurance**: Ensure quality standards are met
3. **Process Improvement**: Continuously improve documentation processes
4. **Tool Management**: Manage documentation tools and resources

### Translators

1. **Translation**: Translate documents accurately and naturally
2. **Review**: Review translations for quality and accuracy
3. **Consistency**: Maintain consistency with existing translations
4. **Feedback**: Provide feedback on the translation process

### Technical Reviewers

1. **Technical Accuracy**: Verify technical content accuracy
2. **Best Practices**: Ensure documentation follows best practices
3. **Completeness**: Ensure documentation is complete and comprehensive
4. **Clarity**: Ensure technical concepts are clearly explained

## Common Issues and Solutions

### Terminology Inconsistency

**Problem**: Technical terms are translated differently in different documents
**Solution**: Maintain a glossary and reference it during translation

### Outdated Translations

**Problem**: One language version is updated but others are not
**Solution**: Implement a process to ensure all language versions are updated together

### Cultural References

**Problem**: Examples or references don't make sense in other cultures
**Solution**: Adapt examples and references for cultural relevance

### Formatting Issues

**Problem**: Formatting is inconsistent between language versions
**Solution**: Use templates and style guides to maintain consistency

## Conclusion

Maintaining high-quality multilingual documentation requires ongoing effort and attention to detail. By following these guidelines, we can ensure that our documentation remains accurate, consistent, and useful for users in all supported languages.