# paper

A dependency-free client for the Dropbox Paper API.

## Installation

Use the `paper` package via `dep`.

```
dep ensure -add github.com/kyleconroy/paper
```

## Usage

```go
package main

import (
	"log"

	"github.com/kyleconroy/paper"
)

func PaperSync(names string) error {
	client := paper.NewClient(os.Getenv("DROPBOX_API_KEY"))
	ctx := context.Background()

	list, err := client.ListDocs(ctx, &paper.ListPaperDocsArgs{Limit: 100})
	if err != nil {
		panic(err)
	}

	for _, doc := range list.DocIDs {
		folder, err := client.GetDocFolderInfo(ctx, &paper.RefPaperDoc{DocID: doc})
		if err != nil {
			panic(err)
		}
		if len(folder.Folders) > 0 {
			log.Printf("Document %s is inside folder %s", doc, folder.Folders[0].Name)
		}
		download, content, err := client.DownloadDoc(ctx, &paper.PaperDocExport{
			DocID: doc,
			Format: paper.ExportFormatMarkdown,
		})
		if err != nil {
			panic(err)
		}
		log.Println(download.Title)
		log.Println(string(content))
	}
}
```
