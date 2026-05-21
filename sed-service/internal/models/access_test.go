package models

import "testing"

func claims(sub string, roles ...string) *Claims {
	return &Claims{Sub: sub, RealmRoles: roles}
}

func doc(author string) *Document {
	return &Document{AuthorSub: author}
}

func dt(readers, writers []string) *DocumentType {
	return &DocumentType{ReaderRoles: readers, WriterRoles: writers}
}

func TestCanReadDocument(t *testing.T) {
	pr := dt([]string{RoleDocReadProcurement, RoleDocReadFinance}, []string{RoleDocWriteProcurement})
	so := dt([]string{RoleDocReadSales}, []string{RoleDocWriteSales})

	t.Run("admin sees all", func(t *testing.T) {
		c := claims("u1", RoleSedAdmin)
		if !CanReadDocument(c, doc("other"), pr, false) {
			t.Fatal("admin must read")
		}
	})

	t.Run("author sees own without doc_read", func(t *testing.T) {
		c := claims("author1", RoleSedAuthor)
		if !CanReadDocument(c, doc("author1"), so, false) {
			t.Fatal("author must read own")
		}
		if CanReadDocument(c, doc("other"), so, false) {
			t.Fatal("author must not read other's without role")
		}
	})

	t.Run("finance reads procurement not sales", func(t *testing.T) {
		c := claims("fin", RoleDocReadFinance)
		if !CanReadDocument(c, doc("x"), pr, false) {
			t.Fatal("finance must read PR type")
		}
		if CanReadDocument(c, doc("x"), so, false) {
			t.Fatal("finance must not read SO type")
		}
	})

	t.Run("sales reads sales not procurement", func(t *testing.T) {
		c := claims("sales", RoleDocReadSales)
		if !CanReadDocument(c, doc("x"), so, false) {
			t.Fatal("sales must read SO")
		}
		if CanReadDocument(c, doc("x"), pr, false) {
			t.Fatal("sales must not read PR")
		}
	})

	t.Run("approver pending", func(t *testing.T) {
		c := claims("ap", RoleSedApprover)
		if !CanReadDocument(c, doc("other"), so, true) {
			t.Fatal("approver with pending task must read")
		}
		if CanReadDocument(c, doc("other"), so, false) {
			t.Fatal("approver without pending must not read without reader role")
		}
	})
}

func TestCanCreateDocument(t *testing.T) {
	dt := dt([]string{RoleDocReadProcurement}, []string{RoleDocWriteProcurement})

	t.Run("writer with sed_author", func(t *testing.T) {
		c := claims("a", RoleSedAuthor, RoleDocWriteProcurement)
		if !CanCreateDocument(c, dt) {
			t.Fatal("expected create allowed")
		}
	})

	t.Run("read only denied", func(t *testing.T) {
		c := claims("a", RoleDocReadProcurement)
		if CanCreateDocument(c, dt) {
			t.Fatal("read-only must not create")
		}
	})
}

func TestDefaultACLForDocumentType(t *testing.T) {
	r, w := DefaultACLForDocumentType("PURCHASE_ORDER_APPROVAL", "NONE")
	if len(r) != 2 || r[0] != RoleDocReadProcurement {
		t.Fatalf("readers: %v", r)
	}
	if len(w) != 1 || w[0] != RoleDocWriteProcurement {
		t.Fatalf("writers: %v", w)
	}
	r, w = DefaultACLForDocumentType("SALES_ORDER_APPROVAL", "NONE")
	if len(r) != 1 || r[0] != RoleDocReadSales {
		t.Fatalf("sales readers: %v", r)
	}
	if w[0] != RoleDocWriteSales {
		t.Fatalf("sales writers: %v", w)
	}
}

func TestCanViewSED(t *testing.T) {
	if CanViewSED(claims("v", RoleSedViewer)) {
		t.Fatal("sed_viewer alone must not pass CanViewSED")
	}
	if !CanViewSED(claims("r", RoleDocReadFinance)) {
		t.Fatal("doc_read must pass CanViewSED")
	}
}
