//go:build !solution

package retryupdate

import (
	"errors"
	"fmt"

	"github.com/gofrs/uuid"
	"gitlab.com/slon/shad-go/retryupdate/kvapi"
)

func UpdateValue(c kvapi.Client, key string, updateFn func(oldValue *string) (newValue string, err error)) error {
	var (
		oldValue    *string
		oldVersion  uuid.UUID
		keyNotFound bool

		apiErr      *kvapi.APIError
		authErr     *kvapi.AuthError
		conflictErr *kvapi.ConflictError
	)

	for {
	GetLoop:
		for {
			resp, err := c.Get(&kvapi.GetRequest{Key: key})

			switch {
			case err == nil:
				oldValue = &resp.Value
				oldVersion = resp.Version
				break GetLoop
			case errors.Is(err, kvapi.ErrKeyNotFound):
				keyNotFound = true
				break GetLoop
			case errors.As(err, &apiErr):
				if errors.As(apiErr.Err, &authErr) {
					return apiErr
				}
				continue
			default:
				return fmt.Errorf("update: unexpected error in Get: %w", err)
			}
		}

		var (
			ignoreUpdateFn bool
			newValue       string
			newVersion     uuid.UUID
			err            error
		)

	SetLoop:
		for {
			if keyNotFound {
				oldValue = nil
				oldVersion = uuid.UUID{}
				keyNotFound = false
				ignoreUpdateFn = false
			}

			if !ignoreUpdateFn {
				newValue, err = updateFn(oldValue)
				if err != nil {
					return err
				}

				newVersion = uuid.Must(uuid.NewV4())
				ignoreUpdateFn = true
			}

			_, err = c.Set(&kvapi.SetRequest{
				Key: key, Value: newValue,
				OldVersion: oldVersion, NewVersion: newVersion,
			})

			switch {
			case err == nil:
				return nil
			case errors.Is(err, kvapi.ErrKeyNotFound):
				keyNotFound = true
				continue
			case errors.As(err, &conflictErr):
				if conflictErr.ExpectedVersion == newVersion {
					return nil
				}
				break SetLoop
			case errors.As(err, &apiErr):
				if errors.As(apiErr.Err, &authErr) {
					return apiErr
				}
				continue
			default:
				return fmt.Errorf("update: unexpected error in Set: %w", err)
			}
		}
	}
}
