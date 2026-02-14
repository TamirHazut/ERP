package pipeline

import (
	"go.mongodb.org/mongo-driver/bson"
)

// PipelineBuilder provides a fluent interface for building MongoDB aggregation pipelines
type PipelineBuilder struct {
	stages []bson.M
}

// New creates a new PipelineBuilder
func New() *PipelineBuilder {
	return &PipelineBuilder{
		stages: make([]bson.M, 0),
	}
}

// Match adds a $match stage to filter documents
func (b *PipelineBuilder) Match(filter bson.M) *PipelineBuilder {
	b.stages = append(b.stages, bson.M{"$match": filter})
	return b
}

// Lookup adds a $lookup stage to perform a left outer join
func (b *PipelineBuilder) Lookup(from, localField, foreignField, as string) *PipelineBuilder {
	b.stages = append(b.stages, bson.M{
		"$lookup": bson.M{
			"from":         from,
			"localField":   localField,
			"foreignField": foreignField,
			"as":           as,
		},
	})
	return b
}

// Unwind adds an $unwind stage to deconstruct an array field
func (b *PipelineBuilder) Unwind(path string, preserveNullAndEmpty bool) *PipelineBuilder {
	b.stages = append(b.stages, bson.M{
		"$unwind": bson.M{
			"path":                       path,
			"preserveNullAndEmptyArrays": preserveNullAndEmpty,
		},
	})
	return b
}

// Project adds a $project stage to include/exclude fields
func (b *PipelineBuilder) Project(projection bson.M) *PipelineBuilder {
	b.stages = append(b.stages, bson.M{"$project": projection})
	return b
}

// ProjectFields adds a $project stage with field names (convenience method)
func (b *PipelineBuilder) ProjectFields(fields []string) *PipelineBuilder {
	if len(fields) == 0 {
		return b
	}

	projection := bson.M{}
	for _, field := range fields {
		projection[field] = 1
	}
	return b.Project(projection)
}

// Group adds a $group stage to group documents by an expression
func (b *PipelineBuilder) Group(id interface{}, accumulator bson.M) *PipelineBuilder {
	groupStage := bson.M{"_id": id}
	for key, value := range accumulator {
		groupStage[key] = value
	}
	b.stages = append(b.stages, bson.M{"$group": groupStage})
	return b
}

// Sort adds a $sort stage to order documents
func (b *PipelineBuilder) Sort(sort bson.M) *PipelineBuilder {
	b.stages = append(b.stages, bson.M{"$sort": sort})
	return b
}

// Limit adds a $limit stage to restrict the number of documents
func (b *PipelineBuilder) Limit(limit int64) *PipelineBuilder {
	b.stages = append(b.stages, bson.M{"$limit": limit})
	return b
}

// Skip adds a $skip stage to skip a number of documents
func (b *PipelineBuilder) Skip(skip int64) *PipelineBuilder {
	b.stages = append(b.stages, bson.M{"$skip": skip})
	return b
}

// AddFields adds an $addFields stage to add new fields to documents
func (b *PipelineBuilder) AddFields(fields bson.M) *PipelineBuilder {
	b.stages = append(b.stages, bson.M{"$addFields": fields})
	return b
}

// ReplaceRoot adds a $replaceRoot stage to replace the root document
func (b *PipelineBuilder) ReplaceRoot(newRoot bson.M) *PipelineBuilder {
	b.stages = append(b.stages, bson.M{"$replaceRoot": newRoot})
	return b
}

// UnionWith adds a $unionWith stage to combine documents from multiple collections
func (b *PipelineBuilder) UnionWith(collection string, pipeline []bson.M) *PipelineBuilder {
	b.stages = append(b.stages, bson.M{
		"$unionWith": bson.M{
			"coll":     collection,
			"pipeline": pipeline,
		},
	})
	return b
}

// Count adds a $count stage to count documents
func (b *PipelineBuilder) Count(fieldName string) *PipelineBuilder {
	b.stages = append(b.stages, bson.M{"$count": fieldName})
	return b
}

// Facet adds a $facet stage to process multiple pipelines on the same set of input documents
func (b *PipelineBuilder) Facet(facets bson.M) *PipelineBuilder {
	b.stages = append(b.stages, bson.M{"$facet": facets})
	return b
}

// Custom adds a custom stage to the pipeline
func (b *PipelineBuilder) Custom(stage bson.M) *PipelineBuilder {
	b.stages = append(b.stages, stage)
	return b
}

// Build returns the completed pipeline stages
func (b *PipelineBuilder) Build() []bson.M {
	return b.stages
}

// Reset clears all stages from the pipeline
func (b *PipelineBuilder) Reset() *PipelineBuilder {
	b.stages = make([]bson.M, 0)
	return b
}

// Len returns the number of stages in the pipeline
func (b *PipelineBuilder) Len() int {
	return len(b.stages)
}
