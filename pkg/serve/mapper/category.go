package mapper

import (
	"fmt"

	"jank.com/jank_blog/internal/global"
	category "jank.com/jank_blog/internal/model/category"
)

// GetCategoryByID 根据 ID 查找类目
func GetCategoryByID(id int64) (*category.Category, error) {
	var cat category.Category
	err := global.DB.Where("id = ? AND deleted = ?", id, 0).First(&cat).Error
	if err != nil {
		return nil, err
	}
	return &cat, nil
}

// GetCategoriesByParentID 根据父类目 ID 查找直接子类目
func GetCategoriesByParentID(parentID int64) ([]category.Category, error) {
	var categories []category.Category
	err := global.DB.Where("parent_id = ? AND deleted = ?", parentID, 0).Find(&categories).Error
	if err != nil {
		return nil, err
	}
	return categories, nil
}

// GetCategoriesByParentPath 根据父类目的路径查找所有子类目
func GetCategoriesByParentPath(parentPath string) ([]category.Category, error) {
	var categories []category.Category
	err := global.DB.Where("path LIKE ? AND deleted = ?", fmt.Sprintf("%s%%", parentPath), 0).Find(&categories).Error
	if err != nil {
		return nil, err
	}
	return categories, nil
}

// GetCategoriesByPath 根据路径获取所有子类目
func GetCategoriesByPath(path string) ([]*category.Category, error) {
	var categories []*category.Category
	err := global.DB.Model(&category.Category{}).
		Where("path LIKE ? AND deleted = ?", fmt.Sprintf("%s%%", path), 0).
		Find(&categories).Error
	if err != nil {
		return nil, err
	}

	return categories, nil
}

// GetAllActivatedCategories 获取所有未删除的类目
func GetAllActivatedCategories() ([]*category.Category, error) {
	var categories []*category.Category
	err := global.DB.Where("deleted = ?", 0).
		Find(&categories).Error

	if err != nil {
		return nil, err
	}

	return categories, nil
}

// GetParentCategoryPathByID 根据父类目 ID 查找父类目的路径
func GetParentCategoryPathByID(parentID int64) (string, error) {
	if parentID == 0 {
		return "", nil
	}
	var parentCategory *category.Category
	err := global.DB.Select("path").Where("id = ? AND deleted = ?", parentID, 0).First(&parentCategory).Error
	if err != nil {
		return "", err
	}
	return parentCategory.Path, nil
}

// CreateCategory 将新类目保存到数据库
func CreateCategory(newCategory *category.Category) error {
	return global.DB.Create(newCategory).Error
}

// UpdateCategory 更新类目信息
func UpdateCategory(category *category.Category) error {
	return global.DB.Save(category).Error
}

// DeleteCategoriesByPathSoftly 软删除类目及其子类目
func DeleteCategoriesByPathSoftly(path string) error {
	return global.DB.Model(&category.Category{}).
		Where("path LIKE ? AND deleted = ?", fmt.Sprintf("%s%%", path), 0).
		Update("deleted", 1).Error
}
