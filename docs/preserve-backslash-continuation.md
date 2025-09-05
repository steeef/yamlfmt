# Preserve Backslash Continuation Formatter

## Problem Statement

The basic yamlfmt formatter currently reformats long strings by automatically folding them when they exceed the configured `max_line_length`. This behavior is problematic for strings that are intentionally formatted with backslash continuation across multiple lines.

### Example Issue

**Input (desired format):**
```yaml
metadata:
  annotations:
    external-dns.alpha.kubernetes.io/hostname: "\
      airflow.tataridev.com.,\
      argocd-cli.prod.tatari.dev.,\
      api.tatari.tv.,\
      auth.tatari.tv."
```

**Current yamlfmt output:**
```yaml
metadata:
  annotations:
    external-dns.alpha.kubernetes.io/hostname: "airflow.tataridev.com., argocd-cli.prod.tatari.dev., api.tatari.tv.,
      auth.tatari.tv."
```

**Desired behavior:** Preserve the original backslash-continued format when it's explicitly used.

## Goal

Create a custom formatter that:
1. Detects strings using backslash continuation (`\` at end of lines)
2. Preserves their multi-line format during formatting
3. Applies normal yamlfmt formatting to all other content

## Implementation Attempts

### Attempt 1: YAML AST Level Approach (Current WIP)

**Files created:**
- `formatters/basic/preserve_backslash_formatter.go`
- `formatters/basic/preserve_backslash_factory.go`
- Modified `cmd/yamlfmt/main.go` to register new formatter

**Approach:**
- Extend `BasicFormatter` with custom logic
- Try to detect backslash-continued strings before YAML parsing
- Use custom encoder settings to preserve format

**Issue discovered:**
The YAML parser (go-yaml) collapses backslash-continued strings during the parsing phase, before our custom formatter can preserve them. The strings are already combined when they reach the YAML AST level.

**Test results:**
```bash
# Using preserve_backslash formatter
./yamlfmt -conf test_preserve_config.yaml -dry test_multiline2.yaml
# Shows that backslash continuation is still collapsed
```

### Next Steps

The current AST-level approach won't work because the YAML parser normalizes the strings before we can intervene. Need to implement a **text-level approach**:

1. **Pre-processing phase:**
   - Scan raw text for backslash-continued strings
   - Replace them with placeholder tokens
   - Apply normal yamlfmt formatting to the modified text

2. **Post-processing phase:**
   - Restore original backslash-continued strings from placeholders
   - Maintain proper indentation context

3. **Alternative approach:**
   - Implement a feature that works at the encoder level
   - Override string encoding behavior for detected continuation patterns

## Configuration

The new formatter type can be configured as:

```yaml
formatter:
  type: preserve_backslash
  indent: 2
  include_document_start: false
  max_line_length: 140
  pad_line_comments: 2
  retain_line_breaks: true
```

## Current Workaround

Users currently work around this issue by using `regex_exclude` to skip formatting specific fields:

```yaml
formatter:
  type: basic
  # ... other config
regex_exclude:
  - "external-dns\\.alpha\\.kubernetes\\.io/hostname"
```

This approach excludes the entire field from formatting, which is not ideal.

## References

- yamlfmt basic formatter: `formatters/basic/`
- YAML encoder settings: `formatter.go:104-135`
- Feature system: `formatters/basic/features/`
