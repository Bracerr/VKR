package usecases

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/google/uuid"
)

func parseTimeOrNil(s string) (*time.Time, error) {
	if s == "" {
		return nil, nil
	}
	t, err := time.Parse(time.RFC3339, s)
	if err != nil {
		return nil, err
	}
	return &t, nil
}

func (a *App) Search(ctx context.Context, tenant, serialNo, batchID, productID, fromS, toS string) (map[string]any, error) {
	// MVP: ищем якорные узлы по сериалу (SERIAL meta.serial_no) или по batch_id (узел BATCH по external_id).
	// Фильтры from/to/product_id пока возвращаем как echo, граф фильтрами будет резаться на клиенте/следующей итерации.

	if serialNo == "" && batchID == "" {
		return nil, fmt.Errorf("%w: serial_no или batch_id обязателен", ErrValidation)
	}
	if _, err := parseTimeOrNil(fromS); err != nil {
		return nil, fmt.Errorf("%w: from", ErrValidation)
	}
	if _, err := parseTimeOrNil(toS); err != nil {
		return nil, fmt.Errorf("%w: to", ErrValidation)
	}

	var anchors []map[string]any
	if batchID != "" {
		id, err := a.Store.GetNodeID(ctx, nil, tenant, "BATCH", batchID)
		if err != nil {
			return nil, err
		}
		if id != nil {
			n, err := a.Store.GetNodeByID(ctx, nil, tenant, *id)
			if err != nil {
				return nil, err
			}
			if n != nil {
				anchors = append(anchors, n)
			}
		}
	}
	if serialNo != "" {
		// простой поиск: среди SERIAL узлов где meta.serial_no == serialNo
		rows, err := a.Store.Pool.Query(ctx, `
			SELECT id FROM trace_nodes
			WHERE tenant_code=$1 AND node_type='SERIAL' AND (meta->>'serial_no') = $2
			LIMIT 50
		`, tenant, serialNo)
		if err != nil {
			return nil, err
		}
		defer rows.Close()
		for rows.Next() {
			var id uuid.UUID
			if err := rows.Scan(&id); err != nil {
				return nil, err
			}
			n, err := a.Store.GetNodeByID(ctx, nil, tenant, id)
			if err != nil {
				return nil, err
			}
			if n != nil {
				anchors = append(anchors, n)
			}
		}
	}

	return map[string]any{
		"tenant_code": tenant,
		"filters": map[string]any{
			"serial_no":  serialNo,
			"batch_id":   batchID,
			"product_id": productID,
			"from":       fromS,
			"to":         toS,
		},
		"anchors": anchors,
	}, nil
}

func (a *App) Graph(ctx context.Context, tenant, anchorType, anchorID, fromS, toS, depthS string) (map[string]any, error) {
	if anchorType == "" || anchorID == "" {
		return nil, fmt.Errorf("%w: anchor_type/anchor_id", ErrValidation)
	}
	if _, err := parseTimeOrNil(fromS); err != nil {
		return nil, fmt.Errorf("%w: from", ErrValidation)
	}
	if _, err := parseTimeOrNil(toS); err != nil {
		return nil, fmt.Errorf("%w: to", ErrValidation)
	}
	depth := 2
	if depthS != "" {
		v, err := strconv.Atoi(depthS)
		if err != nil || v < 1 || v > 6 {
			return nil, fmt.Errorf("%w: depth", ErrValidation)
		}
		depth = v
	}

	anchorNodeID := uuid.Nil
	if u, err := uuid.Parse(anchorID); err == nil {
		anchorNodeID = u
	} else {
		// если передали внешний id — ищем по (type, external_id)
		id, err := a.Store.GetNodeID(ctx, nil, tenant, anchorType, anchorID)
		if err != nil {
			return nil, err
		}
		if id == nil {
			return nil, ErrNotFound
		}
		anchorNodeID = *id
	}

	// BFS по trace_edges, без жёсткого фильтра по времени в MVP (мета/time можно добавить позже).
	visited := map[uuid.UUID]struct{}{anchorNodeID: {}}
	frontier := []uuid.UUID{anchorNodeID}
	var edgesOut []map[string]any

	for i := 0; i < depth; i++ {
		if len(frontier) == 0 {
			break
		}
		eds, err := a.Store.ListEdgesTouching(ctx, nil, tenant, frontier)
		if err != nil {
			return nil, err
		}
		next := make([]uuid.UUID, 0, len(frontier)*2)
		for _, e := range eds {
			var meta any
			_ = json.Unmarshal(e.Meta, &meta)
			edgesOut = append(edgesOut, map[string]any{
				"id":          e.ID.String(),
				"edge_type":   e.EdgeType,
				"from_node_id": e.FromNodeID.String(),
				"to_node_id":   e.ToNodeID.String(),
				"meta":        meta,
			})
			for _, nid := range []uuid.UUID{e.FromNodeID, e.ToNodeID} {
				if _, ok := visited[nid]; !ok {
					visited[nid] = struct{}{}
					next = append(next, nid)
				}
			}
		}
		frontier = next
	}

	// nodes
	nodes := make([]map[string]any, 0, len(visited))
	for id := range visited {
		n, err := a.Store.GetNodeByID(ctx, nil, tenant, id)
		if err != nil {
			return nil, err
		}
		if n != nil {
			nodes = append(nodes, n)
		}
	}

	return map[string]any{
		"tenant_code": tenant,
		"anchor": map[string]any{
			"anchor_type": anchorType,
			"anchor_id":   anchorID,
		},
		"filters": map[string]any{
			"from":  fromS,
			"to":    toS,
			"depth": depth,
		},
		"nodes": nodes,
		"edges": edgesOut,
	}, nil
}

