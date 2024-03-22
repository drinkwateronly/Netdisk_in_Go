package test

import (
	"fmt"
	"netdisk_in_go/models"
	ApiModels "netdisk_in_go/models/api_models"
	"testing"
)

//func TestGorm(t *testing.T) {
//	db, err := gorm.Open(mysql.Open("root:19990414@tcp(127.0.0.1:3306)/netdisk?charset=utf8mb4&parseTime=True&loc=Local"), &gorm.config{})
//	if err != nil {
//		panic("failed to connect database")
//	}
//	// 迁移 schema
//	db.AutoMigrate(&models.UserBasic{})
//	db.AutoMigrate(&models.UserRepository{})
//	db.AutoMigrate(&models.RepositoryPool{})
//	db.AutoMigrate(&models.RecoveryBasic{})
//	db.AutoMigrate(&models.ShareRepository{})
//	db.AutoMigrate(&models.ShareBasic{})
//
//	//user := &models.UserBasic{
//	//	Name: "chenjie",
//	//}
//	//// Create
//	//db.Create(user)
//	//
//	//fmt.Println(db.First(&user, 1))
//	//
//	//// Update - 将 product 的 price 更新为 200
//	//db.Model(&user).Update("Password", "990414")
//	//// Update - 更新多个字段
//}
//
//func TestFind(t *testing.T) {
//	common.InitMySQL()
//	ub, isExist := models.FindUserByPhone(models.DB, "18927841103")
//	if isExist {
//		t.Fatal("?")
//	}
//	fmt.Println(ub)
//}
//
//func TestDigui(t *testing.T) {
//	common.InitMySQL()
//	var ur []models.UserRepository
//	res := models.DB.Raw(`with RECURSIVE temp as
//(
//    select * from user_repository where file_name="/"
//    union all
//    select ur.* from user_repository as ur,temp t where ur.parent_id=t.user_file_id and ur.is_dir = 1 AND ur.deleted_at is NULL
//)
//select temp.parent_id, temp.user_file_id, temp.is_dir, temp.file_name from temp;`).Find(&ur)
//	if res.Error != nil {
//		t.Fatal(res.Error)
//		return
//	}
//	fmt.Println(res.RowsAffected, len(ur))
//
//}

func BuildFileTree() (*ApiModels.UserFileTreeNode, error) {
	// 存放查询结果
	var dirs []models.UserRepository
	// 用户一定有个根目录, 从根目录递归mysql查询所有文件夹
	res := models.DB.Raw(`with RECURSIVE temp as
(
    SELECT * from user_repository where file_name="/" AND user_id = '7e848eb2-a569-4a5b-a828-51d985c60896'
    UNION ALL
    SELECT ur.* from user_repository as ur,temp t 
	where ur.parent_id=t.user_file_id and ur.is_dir = 1 AND ur.deleted_at is NULL
)
select * from temp;`).Find(&dirs)
	if res.Error != nil {
		return nil, res.Error
	}
	fmt.Printf("%v\n", dirs)
	// 递归mysql查询结果中，越上层的文件记录越靠前，且同一个父文件夹下的结果都会相邻
	root := ApiModels.UserFileTreeNode{
		UserFileId: dirs[0].UserFileId,
		DirName:    dirs[0].FileName,
		FilePath:   dirs[0].FilePath,
		Depth:      0,
		State:      "closed",
		IsLeaf:     nil,
		Children:   make([]*ApiModels.UserFileTreeNode, 0),
	}
	// 建队，根节点入队
	nodeMaps := make(map[string]*ApiModels.UserFileTreeNode)
	//queue := make([]*ApiModels.UserFileTreeNode, 1)
	nodeMaps[root.UserFileId] = &root
	// 设置为当前节点，创建孩子节点空列表

	//children := make([]*ApiModels.UserFileTreeNode, 0)
	// 存放节点文件路径
	var filePath string
	//curParentId := root.UserFileId
	//curNode := &root
	// 遍历一遍查询结果dirs
	dirLen := len(dirs)
	for i := 1; i < dirLen; i++ {
		// 拼接文件路径
		if dirs[i].FilePath == "/" {
			filePath = "/" + dirs[i].FileName
		} else {
			filePath = dirs[i].FilePath + "/" + dirs[i].FileName
		}
		// 孩子节点
		child := ApiModels.UserFileTreeNode{
			ParentId:   dirs[i].ParentId,
			UserFileId: dirs[i].UserFileId,
			DirName:    dirs[i].FileName,
			FilePath:   filePath,
			Depth:      0,
			State:      "closed",
			IsLeaf:     nil,
			Children:   make([]*ApiModels.UserFileTreeNode, 0),
		}
		fmt.Printf("%v\n", child)

		// 当前根节点的孩子
		//children = append(children, &child)

		nodeMaps[dirs[i].UserFileId] = &child

		// 改节点在未来可能作为其他节点的父亲
		nodeMaps[child.ParentId].Children = append(nodeMaps[child.ParentId].Children, &child)

		// 找完了当前节点的所有孩子
		//curNode.Children = children
		//nodeMaps[curParentId].Children = children
		// 当前节点（队头）出队
		//delete(nodeMaps, curParentId)

		//curNode = nodeMaps[dirs[i].ParentId]
		//curParentId = dirs[i].ParentId
		// 重置孩子节点切片
		//children = make([]*ApiModels.UserFileTreeNode, 0)
		//children = append(children, &child)

	}
	//nodeMaps[curParentId].Children = children
	return &root, nil
}

func TestBuildTree(t *testing.T) {
	BuildFileTree()

}
