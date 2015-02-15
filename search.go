package main

import "github.com/blevesearch/bleve"

func buildIndexMapping() (*bleve.IndexMapping, error) {
	englishTextFieldMapping := bleve.NewTextFieldMapping()
	englishTextFieldMapping.Analyzer = "en"

	keywordFieldMapping := bleve.NewTextFieldMapping()
	keywordFieldMapping.Analyzer = "keyword"

	lineMapping := bleve.NewDocumentMapping()
	lineMapping.AddFieldMappingsAt("nick", keywordFieldMapping)
	lineMapping.AddFieldMappingsAt("text", englishTextFieldMapping)

	indexMapping := bleve.NewIndexMapping()
	indexMapping.AddDocumentMapping("line", lineMapping)

	indexMapping.DefaultAnalyzer = "en"
	return indexMapping, nil
}
