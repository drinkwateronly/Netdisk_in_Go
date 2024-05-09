package test

import (
	"encoding/csv"
	"fmt"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"math/rand"
	"netdisk_in_go/common"
	"netdisk_in_go/models"
	"os"
	"strconv"
	"testing"
	"time"
)

func genFakePhoneNumber() string {
	var phoneNumber string
	// 先选3位
	phonePrefixs := []string{
		"133", "149", "153", "173", "177",
		"180", "181", "189", "199", "130", "131", "132",
		"145", "155", "156", "166", "171", "175", "176", "185", "186", "166", "134", "135",
		"136", "137", "138", "139", "147", "150", "151", "152", "157", "158", "159", "172",
		"178", "182", "183", "184", "187", "188", "198", "170", "171",
	}
	phoneNumber += phonePrefixs[rand.Intn(len(phonePrefixs))]
	// 剩下8位
	for i := 0; i < 8; i++ {
		phoneNumber += strconv.Itoa(rand.Intn(10))
	}
	return phoneNumber
}

func TestGenerateUserCSV(t *testing.T) {
	// 打开文件存放用户信息
	file, err := os.Create("user.csv")
	if err != nil {
		panic(err)
	}
	defer file.Close()
	writer := csv.NewWriter(file)
	defer writer.Flush()
	//
	history := make(map[string]uint8)

	for i := 0; i < 100000; i++ {
		phoneNumber := genFakePhoneNumber()
		userId := common.GenerateUUID()
		rootFileId := common.GenerateUUID()
		for history[phoneNumber] == 1 {
			// 防止随机生成的电话重复
			phoneNumber = genFakePhoneNumber()
		}
		// 记录已生成的电话号码
		history[phoneNumber] = 1
		// 写入文件
		record := []string{phoneNumber, userId, rootFileId}
		err = writer.Write(record)
		if err != nil {
			panic(err)
		}
	}
}

// 插入大量用户信息
func TestCreateFakeUserAccount(t *testing.T) {
	dsn := "root:123@tcp(172.31.226.34:3306)/netdisk?charset=utf8mb4&parseTime=True&loc=Local"
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		//Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		t.Fatal(err)
	}
	// 打开文件
	file, err := os.Open("user.csv")
	if err != nil {
		panic(err)
	}
	reader := csv.NewReader(file)

	begin := time.Now()
	var ubs []models.UserBasic
	var urs []models.UserRepository
	for i := 0; i < 100; i++ {
		ubs = make([]models.UserBasic, 0, 1000)
		urs = make([]models.UserRepository, 0, 1000)
		for j := 0; j < 1000; j++ {
			record, err := reader.Read()
			//fmt.Println(record)
			if err != nil {
				panic(err)
			}

			ubs = append(ubs, models.UserBasic{
				CreatedAt:        time.Now(),
				UpdatedAt:        time.Now(),
				DeletedAt:        gorm.DeletedAt{},
				UserId:           record[1],
				UserType:         0,
				Username:         "TestAccount",
				Password:         "39ad1ef78e0d490d2f26561c0df04ab9",
				Phone:            record[0],
				Email:            "",
				TotalStorageSize: 0,
				StorageSize:      0,
				Salt:             "89820611-c5de-4417-9053-70e0f1ef81b4",
			})

			urs = append(urs, models.UserRepository{
				UserFileId:    record[2],
				FileId:        "",
				UserId:        record[1],
				FilePath:      "",
				ParentId:      common.GenerateUUID(),
				FileName:      "/",
				ExtendName:    "",
				FileType:      5,
				IsDir:         1,
				FileSize:      0,
				ModifyTime:    "2024-5-3 12:00:00",
				UploadTime:    "2024-5-3 12:00:00",
				DeleteBatchId: "",
				CreatedAt:     time.Now(),
				UpdatedAt:     time.Now(),
				DeletedAt:     0,
			})

		}

		res := db.CreateInBatches(ubs, 100)
		if res.Error != nil {
			fmt.Println(res.Error)
		}
		res = db.CreateInBatches(urs, 100)
		if res.Error != nil {
			fmt.Println(res.Error)
		}
		fmt.Println(i*10000, time.Now().Sub(begin))
	}
}

func TestGenerateUserFileCSV(t *testing.T) {
	// 打开文件存放用户信息
	file, err := os.Create("userfile2.csv")
	if err != nil {
		panic(err)
	}
	defer file.Close()
	writer := csv.NewWriter(file)
	defer writer.Flush()
	for i := 0; i < 100000; i++ {
		record := []string{common.GenerateUUID()}
		err = writer.Write(record)
		if err != nil {
			t.Fatal(err)
		}
	}
}

func TestCreateFakeFileRecord(t *testing.T) {
	dsn := "root:123@tcp(172.31.226.34:3306)/netdisk?charset=utf8mb4&parseTime=True&loc=Local"
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		//Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		t.Fatal(err)
	}
	// 打开文件
	userCSV, err := os.Open("user.csv")
	if err != nil {
		t.Fatal(err)
	}
	userReader := csv.NewReader(userCSV)

	fileidCSV, err := os.Open("userfile2.csv")
	if err != nil {
		t.Fatal(err)
	}
	fileIdReader := csv.NewReader(fileidCSV)

	begin := time.Now()
	var urs []models.UserRepository
	for i := 0; i < 100; i++ {
		urs = make([]models.UserRepository, 0, 1000)
		for j := 0; j < 1000; j++ {
			record1, err := userReader.Read()
			record2, err := fileIdReader.Read()
			if err != nil {
				panic(err)
			}
			urs = append(urs, models.UserRepository{
				UserFileId:    record2[0],
				FileId:        "66e006d2-01cf-4b80-886b-6fa67f1760b6",
				UserId:        record1[1],
				FilePath:      "/",
				ParentId:      record1[2], // 根目录id
				FileName:      "markov tutorial",
				ExtendName:    "pdf",
				FileType:      2,
				IsDir:         0,
				FileSize:      298317,
				ModifyTime:    "2024-5-3 12:00:00",
				UploadTime:    "2024-5-3 12:00:00",
				DeleteBatchId: "",
				CreatedAt:     time.Now(),
				UpdatedAt:     time.Now(),
				DeletedAt:     0,
			})
		}

		res := db.CreateInBatches(urs, 100)
		if res.Error != nil {
			fmt.Println(res.Error)
		}

		fmt.Println(i*10000, time.Now().Sub(begin))

	}
}
