# Wails 自动生成的 JavaScript 绑定

> ⚠️ 此目录由 `wails dev` / `wails build` 命令自动生成。
>
> **不要手动修改此目录下的文件**,它们会在每次构建时重新生成。

## 占位说明

当前 `runtime.js`、`go/main/App.js` 等文件为占位 stub,
确保 `pnpm install` / `pnpm build` 不会因为缺失文件而失败。

当你运行 `wails dev` 或 `wails build` 时,Wails 会自动生成真实的绑定文件。

## 文件说明(生成后)

```
wailsjs/
├── go/
│   ├── main/
│   │   └── App.js          # 对应 app.go 中的绑定方法
│   └── ...
├── runtime/
│   ├── runtime.js          # Wails Runtime API(EventsEmit/EventsOn/...)
│   └── ...
└── README.md
```

## 调试技巧

如果你在 IDE 中看到 `wailsjs` 导入错误:
1. 运行 `wails dev` 启动开发模式
2. Wails 会自动生成 `frontend/wailsjs/` 目录
3. IDE 错误将自动消失

或者手动触发:
```bash
wails generate module
```