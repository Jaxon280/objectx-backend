# objectx-backend

## 全体の設計（メモ）

- アプリケーション: written in `Go`
    - `handler`: handling request and sending response
    - `model`: connecting to database 
    - `image`: uploading and getting images
- データベース: `boltdb` (in-memory database)
- レンタルサーバー: `digitalOcean`
- 画像配信サーバー: `cloudinary`
