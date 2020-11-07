# 物体X Backend API
- `Host` : 未定、フロントエンド側と同じドメイン名を取得してから決定した方が良さそう。

## GET /images
- 画像とそれに付随するパラメータを全て取得する用のAPI

#### Parameter

#### Request

#### Response

以下の要素を持つJSONの配列を返す。

- `id` : 画像の固有ID
- `index` : 全体の投稿順
- `sentiment` : 感情の種類。ユーザーが入力してもらう感情の値をバックエンド側で解析し、"sad" "happy" など文字列で値を返す
- `brightness` : 画像の明度
- `created` : 画像の投稿日時（UNIX時間）
- `tone` : 画像の色調。文字列で返す
- `image_url` : 画像へのURL

*Example*
```
[
    {
        "id": "8dae53ae24589c", // string
        "index": 1, // integer
        "sentiment": "happy", // string
        "brightness": 32, // integer
        "created": 1419933529 // integer,
        "tone": "yellow", // string
        "image_url": "https://res.cloudinary.com/demo/image/upload/v143535505/sample.png" // string
    },
    {
        "id": "8daff3ae24589c", // string
        "index": 2, // integer
        "sentiment": "sad", // string
        "brightness": 129, // integer
        "created": 1419990556 // integer,
        "tone": "blue", // string
        "image_url": "https://res.cloudinary.com/demo/image/upload/v143535505/sample.png" // string
    }
]
```

## GET /images/{id}
- 固有IDをもとに、特定の画像とそれに付随するパラメータを取得する用のAPI

#### Parameter
- `id` : 取得する画像の固有IDをパス内に埋め込む

#### Request

#### Response
以下の要素を持つJSONを返す。

- `id` : 画像の固有ID
- `index` : 全体の投稿順
- `sentiment` : 感情の種類。ユーザーが入力してもらう感情の値をバックエンド側で解析し、"sad" "happy" など文字列で値を返す
- `brightness` : 画像の明度
- `created` : 画像の投稿日時（UNIX時間）
- `tone` : 画像の色調。文字列で返す
- `image_url` : 画像へのURL

*Example*
```
{
    "id": "8daff3ae24589c", // string
    "index": 2, // integer
    "sentiment": "sad", // string
    "brightness": 129, // integer
    "created": 1419990556 // integer,
    "tone": "blue", // string
    "image_url": "https://res.cloudinary.com/demo/image/upload/v143535505/sample.png" // string
}
```


## POST /images
- 画像をパラメータと共に投稿する用のAPI
- `Content-Type` : **multipart/form-data**
    - JSONの部分
        - `Content-Type` : **application/json**
        - `name` : **params**
    - 画像の部分
        - `Content-Type` : **image/png**
        - `name` : **image**

#### Parameter

#### Request
-> パラメータのフォーマットが未定なので決定できない
以下のようなJSONと画像 (PNG形式) を送信する。

```
{
    "sentiment": "sad", // string
    "strength": 12 // integer
}
```

#### Response
以下の要素を持つJSONを返す。

- `id` : 画像の固有ID
- `index` : 全体の投稿順
- `sentiment` : 感情の種類。ユーザーが入力してもらう感情の値をバックエンド側で解析し、"sad" "happy" など文字列で値を返す
- `brightness` : 画像の明度
- `created` : 画像の投稿日時（UNIX時間）
- `tone` : 画像の色調。文字列で返す
- `image_url` : 画像へのURL

*Example*
```
{
    "id": "8daff3ae24589c", // string
    "index": 2, // integer
    "sentiment": "sad", // string
    "brightness": 129, // integer
    "created": 1419990556 // integer,
    "tone": "blue", // string
    "image_url": "https://res.cloudinary.com/demo/image/upload/v143535505/sample.png" // string
}
```

## DELETE /images/{id}
- 固有IDをもとに、特定の画像とそれに付随するパラメータデータを削除する用のAPI

#### Parameter
- `id` : 削除する画像の固有IDをパス内に埋め込む

#### Request

#### Response
以下の要素を持つJSONを返す。

- `id` : 削除した画像の固有ID

*Example*
```
{
    "id": "8daff3ae24589c" // string
}
```

## Error Case

ステータスコードが 4xx/5xx の時、通常のレスポンスの代わりにエラー用の JSON を返す。

##### Response

以下の要素を持つJSONの配列を返す。

- `message` : エラーメッセージを表示する。

*Example*
```
{
    "message": "Something went wrong"
}
```
