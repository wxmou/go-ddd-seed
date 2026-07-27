package repo

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/redis/go-redis/v9"

	appRepo "github.com/go-ddd-seed/go-ddd-seed/internal/application/port/repo"
)

// Ensure interface compliance
var _ appRepo.DictCacheRepository = (*DictCacheRepository)(nil)

const dictCacheKey = "dict:cache"

// DictCacheRepository Redis 字典缓存仓储实现
type DictCacheRepository struct {
	rdb      *redis.Client
	readRepo appRepo.DictTypeReadRepository
}

// NewDictCacheRepository 创建字典缓存仓储
func NewDictCacheRepository(rdb *redis.Client, readRepo appRepo.DictTypeReadRepository) *DictCacheRepository {
	return &DictCacheRepository{rdb: rdb, readRepo: readRepo}
}

// WarmUp 全量缓存预热（启动时调用）
// 加载所有已启用的字典类型及其条目，写入 Redis Hash
func (r *DictCacheRepository) WarmUp(ctx context.Context) error {
	// 查询所有启用类型（分页参数 -1 表示全部状态，传大量 limit 查询全部）
	types, _, err := r.readRepo.List(ctx, 0, 10000, 1)
	if err != nil {
		return fmt.Errorf("warmup: failed to list dict types: %w", err)
	}

	for _, dt := range types {
		entries, err := r.readRepo.FindEntriesByTypeCode(ctx, dt.Code)
		if err != nil {
			continue // 跳过异常类型，不影响其它类型
		}

		cacheEntries := r.toCacheEntries(entries)

		if len(cacheEntries) > 0 {
			data, _ := json.Marshal(cacheEntries)
			if err := r.rdb.HSet(ctx, dictCacheKey, dt.Code, string(data)).Err(); err != nil {
				return fmt.Errorf("warmup: failed to set cache for %s: %w", dt.Code, err)
			}
		}
	}

	return nil
}

// GetEntriesByCode 按 typeCode 获取已启用条目（Cache-Aside：优先缓存，miss 则回源 DB 并回写）
func (r *DictCacheRepository) GetEntriesByCode(ctx context.Context, code string) ([]*appRepo.DictEntryEntry, error) {
	// 1. 优先读缓存
	data, err := r.rdb.HGet(ctx, dictCacheKey, code).Bytes()
	if err == nil {
		var entries []*appRepo.DictEntryEntry
		if err := json.Unmarshal(data, &entries); err != nil {
			return nil, err
		}
		return entries, nil
	}

	if err != redis.Nil {
		return nil, err // Redis 异常
	}

	// 2. 缓存 miss，回源 DB
	entryDTOs, err := r.readRepo.FindEntriesByTypeCode(ctx, code)
	if err != nil {
		return nil, err
	}

	// 3. 构造缓存条目
	cacheEntries := r.toCacheEntries(entryDTOs)

	// 4. 回写缓存（best-effort，不阻断主流程）
	if len(cacheEntries) > 0 {
		data, _ := json.Marshal(cacheEntries)
		_ = r.rdb.HSet(ctx, dictCacheKey, code, string(data)).Err()
	}

	return cacheEntries, nil
}

// toCacheEntries 将 DTO 列表转换为缓存条目列表
func (r *DictCacheRepository) toCacheEntries(dtos []*appRepo.DictEntryDTO) []*appRepo.DictEntryEntry {
	entries := make([]*appRepo.DictEntryEntry, 0, len(dtos))
	for _, e := range dtos {
		entries = append(entries, &appRepo.DictEntryEntry{
			Label: e.Label,
			Value: e.Value,
		})
	}
	return entries
}

// RefreshByCode 刷新指定 typeCode 的缓存
func (r *DictCacheRepository) RefreshByCode(ctx context.Context, code string, entries []*appRepo.DictEntryEntry) error {
	if len(entries) == 0 {
		// 没有启用条目，从缓存中删除
		return r.rdb.HDel(ctx, dictCacheKey, code).Err()
	}

	data, err := json.Marshal(entries)
	if err != nil {
		return err
	}

	return r.rdb.HSet(ctx, dictCacheKey, code, string(data)).Err()
}

// DeleteByCode 删除指定 typeCode 的缓存
func (r *DictCacheRepository) DeleteByCode(ctx context.Context, code string) error {
	return r.rdb.HDel(ctx, dictCacheKey, code).Err()
}
