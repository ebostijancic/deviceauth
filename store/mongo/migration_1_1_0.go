// Copyright 2016 Mender Software AS
//
//    Licensed under the Apache License, Version 2.0 (the "License");
//    you may not use this file except in compliance with the License.
//    You may obtain a copy of the License at
//
//        http://www.apache.org/licenses/LICENSE-2.0
//
//    Unless required by applicable law or agreed to in writing, software
//    distributed under the License is distributed on an "AS IS" BASIS,
//    WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//    See the License for the specific language governing permissions and
//    limitations under the License.
package mongo

import (
	"context"
	"time"

	"github.com/Azure/go-autorest/autorest/to"
	"github.com/mendersoftware/go-lib-micro/mongo/migrate"
	ctxstore "github.com/mendersoftware/go-lib-micro/store"
	"github.com/pkg/errors"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"

	"github.com/mendersoftware/deviceauth/model"
)

type migration_1_1_0 struct {
	ms  *DataStoreMongo
	ctx context.Context
}

type device_0_1_0 struct {
	Id          string    `bson:"_id,omitempty"`
	TenantToken string    `bson:"tenant_token,omitempty"`
	PubKey      string    `bson:"pubkey,omitempty"`
	IdData      string    `bson:"id_data,omitempty"`
	Status      string    `bson:"status,omitempty"`
	CreatedTs   time.Time `bson:"created_ts,omitempty"`
	UpdatedTs   time.Time `bson:"updated_ts,omitempty"`
}

type token_0_1_0 struct {
	Id    string `bson:"_id,omitempty"`
	DevId string `bson:"dev_id,omitempty"`
	Token string `bson:"token,omitempty"`
}

func (m *migration_1_1_0) Up(from migrate.Version) error {
	s := m.ms.session.Copy()

	if err := m.ensureIndexes(s); err != nil {
		return errors.Wrap(err, "database indexing failed")
	}

	defer s.Close()

	iter := s.DB(ctxstore.DbFromContext(m.ctx, DbName)).
		C(DbDevicesColl).Find(nil).Iter()

	var olddev device_0_1_0

	for iter.Next(&olddev) {
		// first prepare an auth set

		// reuse device ID as auth set ID
		asetId := olddev.Id

		aset := model.AuthSet{
			Id:                asetId,
			IdData:            olddev.IdData,
			TenantToken:       olddev.TenantToken,
			PubKey:            olddev.PubKey,
			DeviceId:          olddev.Id,
			Status:            olddev.Status,
			Timestamp:         &olddev.UpdatedTs,
			AdmissionNotified: to.BoolPtr(true),
		}

		if err := s.DB(ctxstore.DbFromContext(m.ctx, DbName)).
			C(DbAuthSetColl).Insert(aset); err != nil {
			return errors.Wrapf(err, "failed to insert auth set for device %v",
				olddev.Id)
		}

		// update tokens
		_, err := s.DB(ctxstore.DbFromContext(m.ctx, DbName)).
			C(DbTokensColl).UpdateAll(
			token_0_1_0{
				DevId: olddev.Id,
			},
			bson.M{
				"$set": bson.M{
					// see model.Token for field naming
					"auth_id": asetId,
				},
			})
		if err != nil {
			return errors.Wrapf(err, "failed to update tokens of device %v", olddev.Id)
		}
	}

	if err := iter.Close(); err != nil {
		return errors.Wrap(err, "failed to close DB iterator")
	}

	return nil
}

func (m *migration_1_1_0) ensureIndexes(s *mgo.Session) error {

	// devices collection
	err := s.DB(ctxstore.DbFromContext(m.ctx, DbName)).
		C(DbDevicesColl).EnsureIndex(mgo.Index{
		Unique: true,
		// identity data shall be unique within collection
		Key:        []string{model.DevKeyIdData},
		Name:       indexDevices_IdentityData,
		Background: false,
	})
	if err != nil {
		return err
	}

	// auth requests
	return s.DB(ctxstore.DbFromContext(m.ctx, DbName)).
		C(DbAuthSetColl).EnsureIndex(mgo.Index{
		Unique: true,
		// tuple (device ID,identity, public key) shall be unique within
		// collection
		Key: []string{
			model.AuthSetKeyDeviceId,
			model.AuthSetKeyIdData,
			model.AuthSetKeyPubKey,
		},
		Name:       indexAuthSet_DeviceId_IdentityData_PubKey,
		Background: false,
	})
}

func (m *migration_1_1_0) Version() migrate.Version {
	return migrate.MakeVersion(1, 1, 0)
}
