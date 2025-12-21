package biz

import "context"

type TagUseCase struct {
	repo TagRepo
	tx   Transaction
}

func NewTagUseCase(repo TagRepo, tx Transaction) *TagUseCase {
	return &TagUseCase{repo: repo, tx: tx}
}

// ===========================================
// 删除标签 (DeleteTag)
// ===========================================

func (uc *TagUseCase) DeleteTag(ctx context.Context, id int64) error {
	return uc.tx.ExecTx(ctx, func(txCtx context.Context) error {
		// 1. 检查是否存在
		tag, err := uc.repo.GetByID(txCtx, id)
		if err != nil {
			return err
		}
		if tag == nil {
			return ErrTagNotFound
		}

		// 2. 删除 (Repo 层需要处理级联删除 post_tags 表中的记录)
		return uc.repo.Delete(txCtx, id)
	})
}
