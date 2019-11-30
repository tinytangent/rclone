package policy

import (
	"context"
	"math"
	
	"github.com/rclone/rclone/fs"
	"github.com/rclone/rclone/backend/union/upstream"
)

func init(){
	registerPolicy("eplfs", &EpLfs{})
}

// EpLfs stands for existing path, least free space
// Of all the candidates on which the path exists choose the one with the least free space.
type EpLfs struct {
	EpAll
}

func (p *EpLfs) lfs(upstreams []*upstream.Fs) (*upstream.Fs, error) {
	var minFreeSpace int64 = math.MaxInt64
	var lfsupstream *upstream.Fs
	for _, u := range upstreams {
		space, err := u.GetFreeSpace()
		if err != nil {
			return nil, err
		}
		if space < minFreeSpace {
			minFreeSpace = space
			lfsupstream = u
		}
	}
	if lfsupstream == nil {
		return nil, fs.ErrorObjectNotFound
	}
	return lfsupstream, nil
}

func (p *EpLfs) lfsEntries(entries []upstream.Entry) (upstream.Entry, error) {
	var minFreeSpace int64
	var lfsEntry upstream.Entry
	for _, e := range entries {
		space, err := e.UpstreamFs().GetFreeSpace()
		if err != nil {
			return nil, err
		}
		if space < minFreeSpace {
			minFreeSpace = space
			lfsEntry = e
		}
	}
	return lfsEntry, nil
}

// Action category policy, governing the modification of files and directories
func (p *EpLfs) Action(ctx context.Context, upstreams []*upstream.Fs, path string) ([]*upstream.Fs, error) {
	upstreams, err := p.EpAll.Action(ctx, upstreams, path)
	if err != nil {
		return nil, err
	}
	u, err := p.lfs(upstreams)
	return []*upstream.Fs{u}, err
}

// ActionEntries is ACTION category policy but receving a set of candidate entries
func (p *EpLfs) ActionEntries(entries ...upstream.Entry) ([]upstream.Entry, error) {
	entries, err := p.EpAll.ActionEntries(entries...)
	if err != nil {
		return nil, err
	}
	e, err := p.lfsEntries(entries)
	return []upstream.Entry{e}, err
}

// Create category policy, governing the creation of files and directories
func (p *EpLfs) Create(ctx context.Context, upstreams []*upstream.Fs, path string) ([]*upstream.Fs, error) {
	upstreams, err := p.EpAll.Create(ctx, upstreams, path)
	if err != nil {
		return nil, err
	}
	u, err := p.lfs(upstreams)
	return []*upstream.Fs{u}, err
}

// CreateEntries is CREATE category policy but receving a set of candidate entries
func (p *EpLfs) CreateEntries(entries ...upstream.Entry) ([]upstream.Entry, error) {
	entries, err := p.EpAll.CreateEntries(entries...)
	if err != nil {
		return nil, err
	}
	e, err := p.lfsEntries(entries)
	return []upstream.Entry{e}, err
}

// Search category policy, governing the access to files and directories
func (p *EpLfs) Search(ctx context.Context, upstreams []*upstream.Fs, path string) (*upstream.Fs, error) {
	if len(upstreams) == 0 {
		return nil, fs.ErrorObjectNotFound
	}
	upstreams, err := p.epall(ctx, upstreams, path)
	if err != nil {
		return nil, err
	}
	return p.lfs(upstreams)
}

// SearchEntries is SEARCH category policy but receving a set of candidate entries
func (p *EpLfs) SearchEntries(entries ...upstream.Entry) (upstream.Entry, error) {
	if len(entries) == 0 {
		return nil, fs.ErrorObjectNotFound
	}
	return p.lfsEntries(entries)
}