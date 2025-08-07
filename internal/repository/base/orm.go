package base

import (
	"gorm.io/gorm"
)

type BaseGormRepository[T any] struct {
	DB *gorm.DB
	//Etcd   *clientv3.Client
	//Consul *consulapi.Client
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

// GetByCondition 根据条件获取单条记录
func (r *BaseGormRepository[T]) GetByCondition(condition interface{}, args ...interface{}) (*T, error) {
	var entity T
	if err := r.DB.Where(condition, args...).First(&entity).Error; err != nil {
		return nil, err
	}
	return &entity, nil
}

// ListByCondition 根据条件获取多条记录
func (r *BaseGormRepository[T]) ListByCondition(condition interface{}, args ...interface{}) ([]T, error) {
	var entities []T
	if err := r.DB.Where(condition, args...).Find(&entities).Error; err != nil {
		return nil, err
	}
	return entities, nil
}

// ListPerPage 分页显示MySQL数据
func (r *BaseGormRepository[T]) ListPerPage(page, pageSize int) (*[]T, int64, error) {
	var entity []T
	var total int64
	if err := r.DB.Debug().Model(&entity).Count(&total).Error; err != nil {
		return &entity, 0, err
	}
	if err := r.DB.Debug().Offset((page - 1) * pageSize).Limit(pageSize).Find(&entity).Error; err != nil {
		return &entity, 0, err
	}
	return &entity, total, nil
}

// ListPerPageWithCondition 支持条件查询的分页显示MySQL数据
func (r *BaseGormRepository[T]) ListPerPageWithCondition(page, pageSize int, condition interface{}, args ...interface{}) (*[]T, int64, error) {
	var entity []T
	var total int64

	// 构建查询
	query := r.DB.Model(&entity)
	if condition != nil {
		query = query.Where(condition, args...)
	}

	// 统计总数
	if err := query.Count(&total).Error; err != nil {
		return &entity, 0, err
	}

	// 分页查询
	if err := query.Offset((page - 1) * pageSize).Limit(pageSize).Find(&entity).Error; err != nil {
		return &entity, 0, err
	}

	return &entity, total, nil
}

// List 列出全部MySQL记录
func (r *BaseGormRepository[T]) List() ([]T, error) {
	var entities []T
	if err := r.DB.Find(&entities).Error; err != nil {
		return nil, err
	}
	return entities, nil
}

// Count 统计记录总数
func (r *BaseGormRepository[T]) Count() (int64, error) {
	var entity T
	var count int64
	if err := r.DB.Model(&entity).Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}

// CountByCondition 根据条件统计记录数
func (r *BaseGormRepository[T]) CountByCondition(condition interface{}, args ...interface{}) (int64, error) {
	var entity T
	var count int64
	if err := r.DB.Model(&entity).Where(condition, args...).Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}

// Exists 检查记录是否存在
func (r *BaseGormRepository[T]) Exists(id uint) (bool, error) {
	var entity T
	var count int64
	if err := r.DB.Model(&entity).Where("id = ?", id).Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}

// ExistsByCondition 根据条件检查记录是否存在
func (r *BaseGormRepository[T]) ExistsByCondition(condition interface{}, args ...interface{}) (bool, error) {
	var entity T
	var count int64
	if err := r.DB.Model(&entity).Where(condition, args...).Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}
