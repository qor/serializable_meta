# Serializable Meta

Serializable Meta is used for cases that a model will keep different type data based on its type, it will serialize different type data as JSON and save it into database.

[![GoDoc](https://godoc.org/github.com/qor/serializable_meta?status.svg)](https://godoc.org/github.com/qor/serializable_meta)

### Serializable Model Definition

```go
type QorJob struct {
	gorm.Model
  Name string
	serializable_meta.SerializableMeta // Embed serializable_meta.SerializableMeta to get the serializable feature
}

// Needs method GetSerializableArgumentResource, so `Serializable Meta` could know your saving argument's type
func (qorJob QorJob) GetSerializableArgumentResource() *admin.Resource {
  return jobsArgumentsMap[qorJob.Kind]
}

var jobsArgumentsMap = map[string]*admin.Resource{
  "newsletter": admin.NewResource(&sendNewsletterArgument{}),
  "import_products": admin.NewResource(&importProductArgument{}),
}

type sendNewsletterArgument struct {
  Subject      string
  Content      string `sql:"size:65532"`
}

type importProductArgument struct {
  ProductsCSV media_library.FileSystem
}
```

### Usage

```go
// Save QorJob with argument into database
var qorJob QorJob
qorJob.Name = "sending newsletter"
qorJob.Kind = "newsletter"
qorJob.SetSerializableArgumentValue(&sendNewsletterArgument{
  Subject: "subject",
  Content: "content",
})

db.Create(&qorJob) // will Marshal `sendNewsletterArgument` as json, and save it into database column `value`
// INSERT INTO "qor_jobs" (kind, value) VALUES (`newsletter`, `{"Subject":"subject","Content":"content"}`);

// Get QorJob and its argument from database
var result QorJob
db.First(&result, "name = ?", "sending newsletter")

// Use its argument
var argument = result.GetSerializableArgument(result)
argument.(*sendNewsletterArgument).Subject // "subject"
argument.(*sendNewsletterArgument).Content // "content"
```

## License

Released under the [MIT License](http://opensource.org/licenses/MIT).
