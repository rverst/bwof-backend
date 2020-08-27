package server

import (
	"encoding/json"
	"fmt"
	"github.com/boltdb/bolt"
	"github.com/google/uuid"
	"github.com/rverst/bwof-backend/pkg/helper"
	"github.com/rverst/bwof-backend/pkg/models"
)

var (
	bucketPics  = []byte("pictures")
	bucketInsta = []byte("instagram")
)

func insertNewPicture(p *models.Picture) error {
	err := db.Update(func(tx *bolt.Tx) error {
		b, err := tx.CreateBucketIfNotExists(bucketPics)
		if err != nil {
			return fmt.Errorf("create bucket %s", err)
		}
		buf, err := json.Marshal(p)
		if err != nil {
			return err
		}

		return b.Put(helper.UUIDtoBytes(p.Id), buf)
	})
	return err
}

func insertNewInstagram(i *models.Instagram) error {
	err := db.Update(func(tx *bolt.Tx) error {
		b, err := tx.CreateBucketIfNotExists(bucketInsta)
		if err != nil {
			return fmt.Errorf("create bucket %s", err)
		}
		buf, err := json.Marshal(i)
		if err != nil {
			return err
		}

		return b.Put(helper.UUIDtoBytes(i.Id), buf)
	})
	return err
}

func updatePicture(p *models.Picture) error {
	err := db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(bucketPics)
		buf, err := json.Marshal(p)
		if err != nil {
			return err
		}
		return b.Put(helper.UUIDtoBytes(p.Id), buf)
	})
	return err
}

func updateInstagram(i *models.Instagram) error {
	err := db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(bucketInsta)
		buf, err := json.Marshal(i)
		if err != nil {
			return err
		}
		return b.Put(helper.UUIDtoBytes(i.Id), buf)
	})
	return err
}

func getDbPicture(id uuid.UUID) (pic *models.Picture, err error) {

	err = db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(bucketPics)
		if b == nil {
			return fmt.Errorf("can't open bucket")
		}
		raw := b.Get(helper.UUIDtoBytes(id))
		if raw == nil {
			return fmt.Errorf("not found")
		}
		var p = &models.Picture{}
		err := json.Unmarshal(raw, p)
		pic = p
		return err
	})
	return pic, err
}

func getDbInstagram(id uuid.UUID) (ins *models.Instagram, err error) {

	err = db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(bucketInsta)
		if b == nil {
			return fmt.Errorf("can't open bucket")
		}
		raw := b.Get(helper.UUIDtoBytes(id))
		if raw == nil {
			return fmt.Errorf("not found")
		}
		var i = &models.Instagram{}
		err := json.Unmarshal(raw, i)
		ins = i
		return err
	})
	return ins, err
}

func getDbPictures() ([]models.Picture, error) {

	list := make([]models.Picture, 0)
	err := db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(bucketPics)
		if b == nil {
			return fmt.Errorf("can't open bucket")
		}

		return b.ForEach(func(k, v []byte) error {
			var p = models.Picture{}
			if err := json.Unmarshal(v, &p); err == nil {
				list = append(list, p)
			}
			return nil
		})
	})
	return list, err
}

func getDbInstagrams() ([]models.Instagram, error) {

	list := make([]models.Instagram, 0)
	err := db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(bucketInsta)
		if b == nil {
			return fmt.Errorf("can't open bucket")
		}

		return b.ForEach(func(k, v []byte) error {
			var i = models.Instagram{}
			if err := json.Unmarshal(v, &i); err == nil {
				list = append(list, i)
			}
			return nil
		})
	})
	return list, err
}

func deleteDbPicture(id uuid.UUID) error {
	err := db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(bucketPics)

		return b.Delete(helper.UUIDtoBytes(id))
	})
	return err
}

func deleteDbInstagram(id uuid.UUID) error {
	err := db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(bucketInsta)

		return b.Delete(helper.UUIDtoBytes(id))
	})
	return err
}
