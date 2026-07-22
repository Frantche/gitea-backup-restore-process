package database

import (
	"bytes"
	"strings"
	"testing"
)

func TestRewritePostgresDumpOwners(t *testing.T) {
	input := strings.NewReader(`ALTER TABLE public.repository OWNER TO app;
ALTER SEQUENCE public.repository_id_seq OWNER TO "old-owner";
CREATE TABLE public.repository (id bigint);
`)
	var output bytes.Buffer

	if err := rewritePostgresDumpOwners(input, &output, `gitea"user`); err != nil {
		t.Fatalf("rewritePostgresDumpOwners failed: %v", err)
	}

	got := output.String()
	if strings.Contains(got, "OWNER TO app") || strings.Contains(got, `OWNER TO "old-owner"`) {
		t.Fatalf("old owner remained in output:\n%s", got)
	}
	if count := strings.Count(got, `OWNER TO "gitea""user";`); count != 2 {
		t.Fatalf("rewritten owner count = %d, want 2; output:\n%s", count, got)
	}
}
