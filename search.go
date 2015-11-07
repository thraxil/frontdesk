package main

import (
	"github.com/blevesearch/bleve"
	"github.com/blevesearch/bleve/analysis/analyzers/keyword_analyzer"
)

func buildIndexMapping() (*bleve.IndexMapping, error) {
	englishTextFieldMapping := bleve.NewTextFieldMapping()
	englishTextFieldMapping.Analyzer = "en"

	keywordFieldMapping := bleve.NewTextFieldMapping()
	keywordFieldMapping.Analyzer = keyword_analyzer.Name

	lineMapping := bleve.NewDocumentMapping()
	lineMapping.AddFieldMappingsAt("nick", keywordFieldMapping)
	lineMapping.AddFieldMappingsAt("text", englishTextFieldMapping)

	indexMapping := bleve.NewIndexMapping()
	indexMapping.AddDocumentMapping("line", lineMapping)

	indexMapping.DefaultAnalyzer = "en"
	return indexMapping, nil
}
