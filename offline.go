package elevengo

import (
	"github.com/deadblue/elevengo/internal/web"
	"github.com/deadblue/elevengo/internal/webapi"
)

type OfflineClearFlag int

const (
	OfflineClearDone OfflineClearFlag = iota
	OfflineClearAll
	OfflineClearFailed
	OfflineClearRunning
	OfflineClearDoneAndDelete
	OfflineClearAllAndDelete

	offlineClearFlagMin = OfflineClearDone
	offlineClearFlagMax = OfflineClearAllAndDelete
)

// OfflineTask describe an offline downloading task.
type OfflineTask struct {
	InfoHash string
	Name     string
	Size     int64
	Status   int
	Percent  float64
	Url      string
	FileId   string
}

func (t *OfflineTask) IsRunning() bool {
	return t.Status == 1
}

func (t *OfflineTask) IsDone() bool {
	return t.Status == 2
}

func (t *OfflineTask) IsFailed() bool {
	return t.Status == -1
}

func (a *Agent) offlineInitToken() (err error) {
	qs := web.Params{}.WithNow("_")
	resp := &webapi.OfflineSpaceResponse{}
	if err = a.wc.CallJsonApi(webapi.ApiOfflineSpace, qs, nil, resp); err != nil {
		return
	}
	a.ot.Time = resp.Time
	a.ot.Sign = resp.Sign
	return nil
}

func (a *Agent) offlineCallApi(url string, params web.Params, resp web.ApiResp) (err error) {
	if a.ot.Time == 0 {
		if err = a.offlineInitToken(); err != nil {
			return
		}
	}
	if params == nil {
		params = web.Params{}
	}
	params.WithInt("uid", a.uid).
		WithInt64("time", a.ot.Time).
		With("sign", a.ot.Sign)
	return a.wc.CallJsonApi(url, nil, params.ToForm(), resp)
}

type offlineIterator struct {
	// Total task count
	count int
	// Page index
	pi int
	// Page count
	pc int
	// Page size
	ps int

	// Cached tasks
	tasks []*webapi.OfflineTask
	// Task index
	index int
	// Task size
	size int

	// Update function
	uf func(*offlineIterator) error
}

func (i *offlineIterator) Next() (err error) {
	if i.index += 1; i.index < i.size {
		return nil
	}
	if i.pi >= i.pc {
		return webapi.ErrReachEnd
	}
	// Fetch next page
	i.pi += 1
	return i.uf(i)
}

func (i *offlineIterator) Index() int {
	return (i.pi-1)*i.ps + i.index
}

func (i *offlineIterator) Get(task *OfflineTask) (err error) {
	if i.index >= i.size {
		return webapi.ErrReachEnd
	}
	t := i.tasks[i.index]
	task.InfoHash = t.InfoHash
	task.Name = t.Name
	task.Size = t.Size
	task.Url = t.Url
	task.Status = t.Status
	task.Percent = t.Percent
	task.FileId = t.FileId
	return nil
}

func (i *offlineIterator) Count() int {
	return i.count
}

// OfflineIterate returns an iterator for travelling offline tasks, it will
// return an error if there are no tasks.
func (a *Agent) OfflineIterate() (it Iterator[OfflineTask], err error) {
	oi := &offlineIterator{
		pi: 1,
		uf: a.offlineIterateInternal,
	}
	if err = a.offlineIterateInternal(oi); err == nil {
		it = oi
	}
	return
}

func (a *Agent) offlineIterateInternal(oi *offlineIterator) (err error) {
	form := web.Params{}.
		WithInt("page", oi.pi)
	resp := &webapi.OfflineListResponse{}
	if err = a.offlineCallApi(webapi.ApiOfflineList, form, resp); err != nil {
		return
	}
	oi.pi = resp.PageIndex
	oi.pc = resp.PageCount
	oi.ps = resp.PageSize
	oi.index, oi.size = 0, len(resp.Tasks)
	if oi.size == 0 {
		err = webapi.ErrReachEnd
	} else {
		oi.tasks = make([]*webapi.OfflineTask, 0, oi.size)
		oi.tasks = append(oi.tasks, resp.Tasks...)
	}
	oi.count = resp.TaskCount
	return
}

// OfflineAdd adds an offline task with url, and saves the downloaded files at
// directory whose ID is dirId.
// You can pass empty string as dirId, to save the downloaded files at default
// directory.
func (a *Agent) OfflineAdd(url string, dirId string) (err error) {
	form := web.Params{}.
		With("url", url)
	if dirId != "" {
		form.With("wp_path_id", dirId)
	}
	resp := &webapi.OfflineAddUrlResponse{}
	return a.offlineCallApi(webapi.ApiOfflineAddUrl, form, resp)
}

// OfflineDelete deletes tasks.
func (a *Agent) OfflineDelete(deleteFiles bool, hashes ...string) (err error) {
	if len(hashes) == 0 {
		return
	}
	form := web.Params{}.WithArray("hash", hashes)
	if deleteFiles {
		form.With("flag", "1")
	}
	return a.offlineCallApi(
		webapi.ApiOfflineDelete, form, &webapi.OfflineBasicResponse{})
}

// OfflineClear clears tasks which is in specific status.
func (a *Agent) OfflineClear(flag OfflineClearFlag) (err error) {
	if flag < offlineClearFlagMin || flag > offlineClearFlagMax {
		flag = OfflineClearDone
	}
	form := web.Params{}.
		WithInt("flag", int(flag))
	return a.offlineCallApi(
		webapi.ApiOfflineClear, form, &webapi.OfflineBasicResponse{})
}
