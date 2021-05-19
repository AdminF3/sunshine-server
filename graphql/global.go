package graphql

import "context"

func (r *mutationResolver) AddEurobor(ctx context.Context, value float64) (*Message, error) {
	if err := r.gl.AddEUROBOR(ctx, value); err != nil {
		return msgErr, err
	}

	return msgOK, nil
}
