# 前端 Proto 类型目录

本目录用于存放由后端目录 `backend/proto/` 生成的 TypeScript 类型。

约定：

- 不手写生成物。
- 前端接口入参和返回值优先引用这里的类型。
- 源头只修改 `backend/proto/*.proto`。
- TypeScript 生成工具会在前端工程初始化后接入。

推荐后续生成目标：

```text
frontend/src/shared/proto/
  common_pb.ts
  user_pb.ts
  assessment_pb.ts
  breed_pb.ts
  pet_pb.ts
  care_pb.ts
```
