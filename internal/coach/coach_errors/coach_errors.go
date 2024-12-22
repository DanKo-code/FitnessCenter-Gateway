package coach_errors

import "errors"

var (
	OnlyPhotoOptional = errors.New("only photo is optional when creating coach")
)
