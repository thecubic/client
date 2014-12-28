// Export-Import for RPC stubs

package libkb

import (
	"fmt"
	"github.com/keybase/protocol/go"
	"time"
)

func (sh SigHint) Export() *keybase_1.SigHint {
	return &keybase_1.SigHint{
		RemoteId:  sh.remoteId,
		ApiUrl:    sh.apiUrl,
		HumanUrl:  sh.humanUrl,
		CheckText: sh.checkText,
	}
}

func (l LinkCheckResult) ExportToIdentifyRow(i int) keybase_1.IdentifyRow {
	return keybase_1.IdentifyRow{
		RowId:     i,
		Proof:     ExportRemoteProof(l.link),
		TrackDiff: ExportTrackDiff(l.diff),
	}
}

func (l LinkCheckResult) Export() keybase_1.LinkCheckResult {
	ret := keybase_1.LinkCheckResult{
		ProofId:     l.position,
		ProofStatus: ExportProofError(l.err),
	}
	if l.cached != nil {
		ret.Cached = l.cached.Export()
	}
	if l.diff != nil {
		ret.Diff = ExportTrackDiff(l.diff)
	}
	if l.remoteDiff != nil {
		ret.RemoteDiff = ExportTrackDiff(l.remoteDiff)
	}
	if l.hint != nil {
		ret.Hint = l.hint.Export()
	}
	return ret
}

func (cr CheckResult) Export() *keybase_1.CheckResult {
	return &keybase_1.CheckResult{
		ProofStatus:   ExportProofError(cr.Status),
		Timestamp:     int(cr.Time.Unix()),
		DisplayMarkup: cr.ToDisplayString(),
	}
}

func ExportRemoteProof(p RemoteProofChainLink) keybase_1.RemoteProof {
	k, v := p.ToKeyValuePair()
	return keybase_1.RemoteProof{
		ProofType:     p.GetIntType(),
		Key:           k,
		Value:         v,
		DisplayMarkup: v,
	}
}

func (i IdentifyRes) ExportToIdentifyOutcome() (res keybase_1.IdentifyOutcome) {
	res.NumTrackFailures = i.NumTrackFailures()
	res.NumTrackChanges = i.NumTrackChanges()
	res.NumProofFailures = i.NumProofFailures()
	res.NumDeleted = i.NumDeleted()
	res.NumProofSuccesses = i.NumProofSuccesses()
	return
}

func (i IdentifyRes) ExportToUncheckedIdentity() (res keybase_1.Identity) {
	res.Status = ExportErrorAsStatus(i.Error)
	if i.TrackUsed != nil {
		res.WhenLastTracked = int(i.TrackUsed.GetCTime().Unix())
	}
	res.Proofs = make([]keybase_1.IdentifyRow, len(i.ProofChecks))
	for j, p := range i.ProofChecks {
		res.Proofs[j] = p.ExportToIdentifyRow(j)
	}
	res.Deleted = make([]keybase_1.TrackDiff, len(i.Deleted))
	for j, d := range i.Deleted {
		// Should have all non-nil elements...
		res.Deleted[j] = *ExportTrackDiff(d)
	}

	return
}

type ExportableError interface {
	error
	ToStatus() keybase_1.Status
}

func ExportProofError(pe ProofError) (ret keybase_1.ProofStatus) {
	if pe == nil {
		ret.State = PROOF_STATE_OK
		ret.Status = PROOF_OK
	} else {
		ret.Status = int(pe.GetStatus())
		ret.State = ProofErrorToState(pe)
		ret.Desc = pe.GetDesc()
	}
	return
}

func ImportProofError(e keybase_1.ProofStatus) ProofError {
	ps := ProofStatus(e.Status)
	if ps == PROOF_STATE_OK {
		return nil
	} else {
		return NewProofError(ps, e.Desc)
	}
}

func ExportErrorAsStatus(e error) (ret keybase_1.Status) {
	if e == nil {
		ret = keybase_1.Status{
			Name: "OK",
			Code: 0,
		}
	} else if ee, ok := e.(ExportableError); ok {
		ret = ee.ToStatus()
	} else {
		ret = keybase_1.Status{
			Name: "GENERIC",
			Code: SC_GENERIC,
			Desc: e.Error(),
		}
	}
	return
}

//=============================================================================

func ImportStatusAsError(s keybase_1.Status) error {
	if s.Code == SC_OK {
		return nil
	} else if s.Code == SC_GENERIC {
		return fmt.Errorf(s.Desc)
	} else {
		ase := AppStatusError{
			Code:   s.Code,
			Name:   s.Name,
			Desc:   s.Desc,
			Fields: make(map[string]bool),
		}
		for _, f := range s.Fields {
			ase.Fields[f] = true
		}
		return ase
	}
}

//=============================================================================

func (a AppStatusError) ToStatus() keybase_1.Status {
	var fields []string
	for k, _ := range a.Fields {
		fields = append(fields, k)
	}

	return keybase_1.Status{
		Code:   a.Code,
		Name:   a.Name,
		Desc:   a.Desc,
		Fields: fields,
	}
}

//=============================================================================

func ExportTrackDiff(d TrackDiff) (res *keybase_1.TrackDiff) {
	if d != nil {
		res = &keybase_1.TrackDiff{
			Type:          keybase_1.TrackDiffType(d.GetTrackDiffType()),
			DisplayMarkup: d.ToDisplayString(),
		}
	}
	return
}

//=============================================================================

func ImportPgpFingerprint(f keybase_1.FOKID) (ret *PgpFingerprint) {
	if f.PgpFingerprint != nil && len(*f.PgpFingerprint) == PGP_FINGERPRINT_LEN {
		var tmp PgpFingerprint
		copy(tmp[:], (*f.PgpFingerprint)[:])
		ret = &tmp
	}
	return
}

func (f *PgpFingerprint) ExportToFOKID() (ret keybase_1.FOKID) {
	slc := (*f)[:]
	ret.PgpFingerprint = &slc
	return
}

//=============================================================================

func (s TrackSummary) Export() (ret keybase_1.TrackSummary) {
	ret.Time = int(s.time.Unix())
	ret.IsRemote = s.isRemote
	return
}

func ImportTrackSummary(s *keybase_1.TrackSummary) *TrackSummary {
	if s == nil {
		return nil
	} else {
		return &TrackSummary{
			time:     time.Unix(int64(s.Time), 0),
			isRemote: s.IsRemote,
		}
	}
}

func ExportTrackSummary(l *TrackLookup) *keybase_1.TrackSummary {
	if l == nil {
		return nil
	} else {
		tmp := l.ToSummary().Export()
		return &tmp
	}
}

//=============================================================================

func (ir *IdentifyRes) Export() *keybase_1.IdentifyOutcome {
	v := make([]string, len(ir.Warnings))
	for i, w := range ir.Warnings {
		v[i] = w.Warning()
	}
	del := make([]keybase_1.TrackDiff, 0, len(ir.Deleted))
	for i, d := range ir.Deleted {
		del[i] = *ExportTrackDiff(d)
	}
	ret := &keybase_1.IdentifyOutcome{
		Status:            ExportErrorAsStatus(ir.Error),
		Warnings:          v,
		TrackUsed:         ExportTrackSummary(ir.TrackUsed),
		NumTrackFailures:  ir.NumTrackFailures(),
		NumTrackChanges:   ir.NumTrackChanges(),
		NumProofFailures:  ir.NumProofFailures(),
		NumDeleted:        ir.NumDeleted(),
		NumProofSuccesses: ir.NumProofSuccesses(),
		Deleted:           del,
	}
	return ret
}

//=============================================================================

func ImportFinishAndPromptRes(f keybase_1.FinishAndPromptRes) (ti TrackInstructions, err error) {
	ti.Local = f.TrackLocal
	ti.Remote = f.TrackRemote
	err = ImportStatusAsError(f.Status)
	return
}

//=============================================================================

func ImportWarnings(v []string) Warnings {
	w := make([]Warning, len(v))
	for i, s := range v {
		w[i] = StringWarning(s)
	}
	return Warnings{w}
}

//=============================================================================
