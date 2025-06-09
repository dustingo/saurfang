package base

import (
	"gorm.io/gorm"
)

type BaseGormRepository[T any] struct {
	DB *gorm.DB
}

// Create创建MySQL数据
func (r *BaseGormRepository[T]) Create(entity *T) error {
	return r.DB.Create(entity).Error
}

// Update更新MYSQL数据
func (r *BaseGormRepository[T]) Update(id uint, entity *T) error {
	return r.DB.Model(entity).Where("id = ?", id).Updates(entity).Error
}

// UpdateALL 全部更新MySQL记录
func (r *BaseGormRepository[T]) UpdateALL(entity *T) error {
	return r.DB.Save(entity).Error
}

// ChangeGroup
// 专用：修改服务器归属
func (r *BaseGormRepository[T]) ChangeGroup(hostID, groupID uint) error {
	var entity T
	return r.DB.Model(&entity).Where("id = ?", hostID).Update("group_id", groupID).Error
}

// delete 删除MySQL记录
func (r *BaseGormRepository[T]) Delete(id uint) error {
	var entity T
	return r.DB.Model(&entity).Where("id = ?", id).Delete(&entity).Error
}

// BatchDelete 批量删除MySQL数据
func (r *BaseGormRepository[T]) BatchDelete(ids []uint) error {
	var entity T
	return r.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("id IN ? ", ids).Delete(&entity).Error; err != nil {
			return err
		}
		return nil
	})
}

// GetByID 根据ID列出MySQL数据
func (r *BaseGormRepository[T]) ListByID(id uint) (*T, error) {
	var entity T
	if err := r.DB.First(&entity, id).Error; err != nil {
		return &entity, err
	}
	return &entity, nil
}

// ListPerPage 分页显示MySQL数据
func (r *BaseGormRepository[T]) ListPerPage(page, pageSize int) (*[]T, int64, error) {
	var entity []T
	var total int64
	if err := r.DB.Model(&entity).Count(&total).Error; err != nil {
		return &entity, 0, err
	}
	if err := r.DB.Offset((page - 1) * pageSize).Limit(pageSize).Find(&entity).Error; err != nil {
		return &entity, 0, err
	}
	return &entity, total, nil
}

// List 列出全部MySQL记录
func (r *BaseGormRepository[T]) List() (*[]T, error) {
	var entity []T
	if err := r.DB.Find(&entity).Error; err != nil {
		return &entity, err
	}
	return &entity, nil
}
