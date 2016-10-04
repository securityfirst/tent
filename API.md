# Octo API

## General

### http Methods

- GET: View
- POST: Create
- PUT: Update
- DELETE: Delete

### Error Response
```
{
	"error": "message with details"
}
```

## Categories

### List
**GET** /api/repo/ (200)

**Sample Response**:```{
	"categories": [
		"catid_1",
		"catid_2"
	]
}```

### Details
**GET** /api/repo/category/:category _(200 - 404)_

**Sample Response**:```{
	"name": "Category name",
	"subcategories": [
		"subid_1",
		"subid_2"
	]
}```

### Create
**POST** /api/repo/category/:category _(201 - 409, 503)_

**Request Body**:```{
	"name": "Category name"
}```

### Update
**PUT** /api/repo/category/:category _(204 - 503)_

**Request Body**:```{
	"name": "Category name"
}```

### Delete
**DELETE** /api/repo/category/:category _(204 - 503)_

## Subcategories

### Details
**GET** /api/repo/category/:category/:sub _(200 - 404)_

**Sample Response**:```{
	"name": "Subcategory name",
	"items": [
		"itemid_1",
		"itemid_2"
	]
}```

### Create
**POST** /api/repo/category/:category/:sub _(201 - 409, 503)_

**Request Body**:```
{
	"name": "Subcategory name"
}```

### Update
**PUT** /api/repo/category/:category/:sub _(204 - 503)_

**Request Body**:```
{
	"name": "Category name"
}```

### Delete
**DELETE** /api/repo/category/:category/:sub _(204 - 503)_

## Items

### Details
**GET** /api/repo/category/:category/:sub/item/:item _(200 - 404)_

**Sample Response**:
```
{
	"hash": "sha1",
	"title": "Item Title",
	"body": "<h1>Sample Body</h1><p>some text</p>",
	"difficulty": "Beginner"
}```

### Create
**POST** /api/repo/category/:category/:sub/item/:item _(201 - 503)_

**Request Body**:
```
{
	"title": "Item Title",
	"body": "<h1>Sample Body</h1><p>some text</p>",
	"difficulty": "Beginner"
}```

### Update
**PUT** /api/repo/category/:category/:sub/item/:item _(204 - 409, 503)_

**Request Body**:
```
{
	"hash": "sha1",
	"title": "Item Title",
	"body": "<h1>Sample Body</h1><p>some text</p>",
	"difficulty": "Beginner"
}```


### Delete
**DELETE** /api/repo/category/:category/:sub/item/:item _(204 - 503)_