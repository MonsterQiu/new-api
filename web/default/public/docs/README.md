# 独立使用文档

- 文档正文：`index.html`
- 页面样式：`assets/styles.css`
- 交互逻辑：`assets/app.js`
- 文档图片：`assets/images/`
- 部署后地址：`https://你的域名/docs/`

修改正文时尽量保留 `id`、`data-search-section` 和侧栏链接中的 `data-section`，这样搜索、目录高亮和移动端导航可以继续正常工作。

编辑完成后，在 `web/default/` 目录重新执行前端构建并部署。
