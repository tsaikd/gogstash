gogstash input lorem
====================

## Synopsis

```yaml
input:
  # type Must be "lorem"
  - type: "lorem"

    # worker count to generate lorem, default: 1
    worker: 1

    # duration to generate lorem, set 0 to generate forever, default: 30s
    duration: "30s"

    # (optional) message format in go text/template
    # support functions:
    #   `TimeFormat(layout string) string`
    #   `Word(min, max int) string`
    #   `Sentence(min, max int) string`
    #   `Paragraph(min, max int) string`
    #   `Email() string`
    #   `Host() string`
    #   `Url() string`
    # eg. '{{.TimeFormat "20060102-150405.000"}}|ERROR|{{.Sentence 1 5}}' =>
    #   '20180921-173749.186|ERROR|Valetudinis tria cura cognitionis.'
    format: '{{.Sentence 1 5}}'

    # (optional) send empty messages without any lorem text, default: false
    empty: false

    # (optional) event extra fields, default: nil
    #fields:
```
