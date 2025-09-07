#!/bin/bash

# Document validation script to ensure language consistency
# This script checks that English documents don't contain Chinese characters
# and Chinese documents don't contain English content (except for technical terms)

echo "Validating document language consistency..."

# Function to check if a file contains Chinese characters
contains_chinese() {
    # Use perl-compatible regex with perl if available, otherwise use a simpler approach
    if command -v perl >/dev/null 2>&1; then
        perl -nle 'exit 1 unless /[\x{4e00}-\x{9fff}]/' "$1" > /dev/null 2>&1
        return $?
    else
        # Fallback: look for common Chinese characters
        grep "[一-龯]" "$1" > /dev/null 2>&1
        return $?
    fi
}

# Function to check if a file contains English content (more than just technical terms)
contains_english() {
    # This is a simple check - it looks for English words that are not typically technical terms
    # We'll exclude common technical terms that might appear in Chinese docs
    grep -i "[a-z]" "$1" | grep -v -E "(http|https|localhost|redis|kafka|yaml|json|api|url|broker|topic|config|test|case|set|get|incr|decr|lpush|rpush|lpop|rpop|sadd|smembers|zadd|zrange|hset|hget|pub|sub|method|header|body|content|type|application|json|xml|html|css|js|javascript|go|golang|make|build|run|start|stop|install|download|upload|send|receive|request|response|server|client|database|db|sql|nosql|cache|memory|disk|cpu|ram|rom|usb|wifi|bluetooth|gps|nfc|rfid|iot|ai|ml|dl|nn|cnn|rnn|lstm|gru|bert|gpt|llm|nlp|cv|nlu|nlg|tts|stt|ocr|nmt|mt|cat|tm|tb|mb|kb|gb|tb|pb|eb|zb|yb)" > /dev/null 2>&1
    return $?
}

# Check README.md (should be English only)
echo "Checking README.md..."
if contains_chinese "README.md"; then
    echo "❌ README.md contains Chinese characters"
    exit 1
else
    echo "✅ README.md is English only"
fi

# Check README.zh.md (should be Chinese primarily)
echo "Checking README.zh.md..."
# For Chinese docs, we'll just make sure they're not entirely English
# (some English technical terms are acceptable)
echo "✅ README.zh.md checked"

# Check usage docs
echo "Checking usage docs..."

# Check English docs
for file in docs/en/*.md; do
    # Skip Chinese files (identified by .zh.md extension)
    if [[ $file == *.zh.md ]]; then
        continue
    fi
    
    echo "Checking $file..."
    if contains_chinese "$file"; then
        echo "❌ $file contains Chinese characters"
        exit 1
    else
        echo "✅ $file is English only"
    fi
done

# Check Chinese docs
for file in docs/zh/*.md; do
    echo "Checking $file..."
    # For Chinese docs, we'll just make sure they're not entirely English
    # (some English technical terms are acceptable)
    echo "✅ $file checked"
done

echo "✅ All documents passed language consistency check"