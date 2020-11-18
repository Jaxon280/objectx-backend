# 物体X Backend API
- `Host` : `http://157.230.43.94` (テスト用)

## GET /images
- 画像とそれに付随するパラメータを全て取得する用のAPI

#### Parameter

#### Request

#### Response

画像とその関連データをまとめたデータの固有IDを Key、と以下の要素を持つ JSON を Value に持つ Key-Value のリストを JSON で返す。
また、アクセス負荷の軽減のために最大でも 40 個の Key-Value を読み込むように設定している。

- `id` : 画像の固有ID
- `sentiment` : 感情の種類。`[-1,1]` の範囲で値を返す
- `brightness` : 画像の明度 ( `HSV` のうち `Value` ) を返す。`Value` は `0-255` の範囲なので、画像の平均の明度が分かる。
- `tone` : 画像の色相 ( `HSV` のうち `Hue` ) を返す。`Hue` は `0-360` の範囲なので、色見本と対応させて画像の平均の色調が分かる。
- `created` : 画像の投稿日時（UNIX時間）
- `image_url` : 画像へのURL

*Example*
```
{
    "6cb9a97c-4a7f-4225-ae04-c15e18c8caed": {
        "id": "6cb9a97c-4a7f-4225-ae04-c15e18c8caed",
        "sentiment": 0.3,
        "brightness": 148.59068468585255,
        "tone": 104.2402622936652,
        "created": 1605725152,
        "image_url": "https://objectx.ams3.cdn.digitaloceanspaces.com/245435e0-2dc1-4b52-b96a-b7db28d73303.png"
    },
    "ae7d7706-55b3-44da-ae5c-093e403fe1d8": {
        "id": "ae7d7706-55b3-44da-ae5c-093e403fe1d8",
        "sentiment": 0.3,
        "brightness": 148.59068468585255,
        "tone": 104.2402622936652,
        "created": 1605721379,
        "image_url": "https://objectx.ams3.cdn.digitaloceanspaces.com/2ac440b2-40cb-4072-8f64-fd1f853fe630.png"
    }
}
```

## GET /images/{id}
- 固有IDをもとに、特定の画像とそれに付随するパラメータを取得する用のAPI

#### Parameter
- `id` : 取得する画像の固有IDをパス内に埋め込む

#### Request

#### Response
以下の要素を持つJSONを返す。

- `id` : 画像の固有ID
- `sentiment` : 感情の種類。`[-1,1]` の範囲で値を返す
- `brightness` : 画像の明度 ( `HSV` のうち `Value` ) を返す。`Value` は `0-255` の範囲なので、画像の平均の明度が分かる。
- `tone` : 画像の色相 ( `HSV` のうち `Hue` ) を返す。`Hue` は `0-360` の範囲なので、色見本と対応させて画像の平均の色調が分かる。
- `created` : 画像の投稿日時（UNIX時間）
- `image_url` : 画像へのURL

*Example*
```
{
    "id": "98850a82-01e7-46aa-aa74-71637d53ab62",
    "sentiment": 0.3,
    "brightness": 148.59068468585255,
    "tone": 104.2402622936652,
    "created": 1605721308,
    "image_url": "https://objectx.ams3.cdn.digitaloceanspaces.com/3779cb61-ff79-4665-b3b2-9acf3af41145.png"
}
```


## POST /upload
- 画像をパラメータと共に投稿する用のAPI

#### Parameter
- `Content-Type` : **multipart/form-data**
    - 感情パラメータの部分
        - `name` : **sentiment**
    - 画像ファイルの部分
        - `name` : **image**

#### Request

#### Response
以下の要素を持つJSONを返す。

- `id` : 画像の固有ID
- `sentiment` : 感情の種類。`[-1,1]` の範囲で値を返す
- `brightness` : 画像の明度 ( `HSV` のうち `Value` ) を返す。`Value` は `0-255` の範囲なので、画像の平均の明度が分かる。
- `tone` : 画像の色相 ( `HSV` のうち `Hue` ) を返す。`Hue` は `0-360` の範囲なので、色見本と対応させて画像の平均の色調が分かる。
- `created` : 画像の投稿日時（UNIX時間）
- `image_url` : 画像へのURL

*Example*
```
{
    "id": "b2627085-1f22-4372-8494-fa3830271ae0",
    "sentiment": 0.3,
    "brightness": 148.59068468585255,
    "tone": 104.2402622936652,
    "created": 1605727252,
    "image_url": "https://objectx.ams3.cdn.digitaloceanspaces.com/819687d5-0b99-4339-af73-810258d63c37.png"
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
    "id": "98850a82-01e7-46aa-aa74-71637d53ab62"
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
    "message": "http: no such file"
}
```
