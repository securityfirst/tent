## Get categories
**URL**: /api/repo/

**Method**: GET

**Success**: Status 200
```
{
	"success": true,
	"categories": [
		"catgeory1",
		"catgeory2"
	]
}
```

## Get category items
**URL**: /api/repo/:category

**Method**: GET

**Success**: Status 200
```
{
	"success": true,
	"items": [
		"catgeory1",
		"catgeory2"
	]
}
```
**Failure**: Status 404
```
{
	"success": false,
	"error": "category not found"
}
```

## Get file contents
**URL**: /api/repo/:category/:item

**Method**: GET

**Success**: Status 200
```
{
	"success": true,
	"hash": "sha1",
	"contents": "file contents"
}
```
**Failure**: Status 404
```
{
	"success": false,
	"error": "file not found"
}
```

## Update file contents
**URL**: /api/repo//:category/:item

**Method**: POST

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
	"success": true
}
```
**Failure**: Status 409
```
{
	"success": false,
	"error": "hash collision"
}
```