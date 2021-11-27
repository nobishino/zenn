---
title:"型セットと「実装」概念"
---

この章が仕様説明編の最後の章となります。ゴールは、次のようなインタフェース型がどのような型によって実装されるかわかることです。

```go
type I interface {
    ~int | string
    fmt.Stringer
    Print() string
}
```

## 