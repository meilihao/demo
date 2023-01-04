# pdf
没选用[unipdf](https://github.com/unidoc/unipdf)原因: unipdf需要[license](https://github.com/unidoc/unipdf-examples/blob/master/README.md)

这里选用了[johnfercher/maroto](https://github.com/johnfercher/maroto), 它借鉴了bootstrap 的网格模式, 使用了gofpdf 生成pdf, 是一个很不错的golang pdf 工具.

## example
- [table](https://github.com/johnfercher/maroto/tree/master/internal/examples/billing)

    支持中文