## Get categories
**URL**: /api/repo/

**Method**: GET

**Success**: Status 200
```
{
	"categories": [
		"category1",
		"category2"
	]
}
```

## Get category items
**URL**: /api/repo/category/:category

**Method**: GET

**Success**: Status 200
```
{
	"items": [
		"item1",
		"item2"
	]
}
```
**Failure**: Status 404
```
{
	"error": "category not found"
}
```

## Get file contents
**URL**: /api/repo/category/:category/item/:item

**Method**: GET

**Success**: Status 200
```
{
	"hash": "sha1",
	"contents": "file contents"
}
```
**Failure**: Status 404
```
{
	"error": "file not found"
}
```

## Update file contents
**URL**: /api/repo/category/:category/item/:item

**Method**: PUT

**Request Body**:
```
{
	"hash": "sha1",
	"contents": "file contents"
}
```
**Success**: Status 200
```
{
	"id":id,
	"hash": "sha1",
	"contents": "file contents"
}
```
**Failure**: Status 409
```
{
	"error": "hash collision"
}
```