//                           _       _
// __      _____  __ ___   ___  __ _| |_ ___
// \ \ /\ / / _ \/ _` \ \ / / |/ _` | __/ _ \
//  \ V  V /  __/ (_| |\ V /| | (_| | ||  __/
//   \_/\_/ \___|\__,_| \_/ |_|\__,_|\__\___|
//
//  Copyright © 2016 - 2019 Weaviate. All rights reserved.
//  LICENSE: https://github.com/semi-technologies/weaviate/blob/develop/LICENSE.md
//  DESIGN & CONCEPT: Bob van Luijt (@bobvanluijt)
//  CONTACT: hello@semi.technology
//

// Package kinds provides managers for all kind-related items, such as things
// and actions. Manager provides methods for "regular" interaction, such as
// add, get, delete, update, etc. Additionally BatchManager allows for
// effecient batch-adding of thing/action instances and references.
package kinds

import (
	"context"
	"fmt"
	"log"

	"github.com/go-openapi/strfmt"
	uuid "github.com/satori/go.uuid"
	"github.com/semi-technologies/weaviate/entities/models"
	"github.com/semi-technologies/weaviate/usecases/config"
	"github.com/semi-technologies/weaviate/usecases/network/common/peers"
	"github.com/sirupsen/logrus"
)

// Manager manages kind changes at a use-case level, i.e. agnostic of
// underlying databases or storage providers
type Manager struct {
	network       network
	config        *config.WeaviateConfig
	repo          Repo
	locks         locks
	schemaManager schemaManager
	logger        logrus.FieldLogger
	authorizer    authorizer
	vectorizer    Vectorizer
	vectorRepo    VectorRepo
}

// Repo describes the requirements the kinds UC has to the connected database
type Repo interface {
	addRepo
	getRepo
	updateRepo
	deleteRepo
	batchRepo
}

type Vectorizer interface {
	Thing(ctx context.Context, concept *models.Thing) ([]float32, error)
	Action(ctx context.Context, concept *models.Action) ([]float32, error)
}

type locks interface {
	LockConnector() (func() error, error)
	LockSchema() (func() error, error)
}

type authorizer interface {
	Authorize(principal *models.Principal, verb, resource string) error
}

type network interface {
	ListPeers() (peers.Peers, error)
}

type VectorRepo interface {
	PutThing(ctx context.Context, concept *models.Thing, vector []float32) error
	PutAction(ctx context.Context, concept *models.Action, vector []float32) error

	DeleteAction(ctx context.Context, className string, id strfmt.UUID) error
	DeleteThing(ctx context.Context, className string, id strfmt.UUID) error
}

// NewManager creates a new manager
func NewManager(repo Repo, locks locks, schemaManager schemaManager,
	network network, config *config.WeaviateConfig, logger logrus.FieldLogger,
	authorizer authorizer, vectorizer Vectorizer, vectorRepo VectorRepo) *Manager {
	return &Manager{
		network:       network,
		config:        config,
		repo:          repo,
		locks:         locks,
		schemaManager: schemaManager,
		logger:        logger,
		vectorizer:    vectorizer,
		authorizer:    authorizer,
		vectorRepo:    vectorRepo,
	}
}

type unlocker interface {
	Unlock() error
}

func unlock(l unlocker) {
	err := l.Unlock()
	if err != nil {
		log.Fatal(err)
	}
}

func generateUUID() (strfmt.UUID, error) {
	uuid, err := uuid.NewV4()
	if err != nil {
		return "", fmt.Errorf("could not generate uuid v4: %v", err)
	}

	return strfmt.UUID(fmt.Sprintf("%v", uuid)), nil
}