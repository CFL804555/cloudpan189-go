// Copyright (c) 2020 tickstep.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
package command

import (
	"fmt"
	"github.com/tickstep/cloudpan189-api/cloudpan"
	"github.com/tickstep/cloudpan189-go/cmder/cmdtable"
	"github.com/tickstep/library-go/logger"
	"os"
	"path"
	"strconv"
	"time"
)

// RunRemove 执行 批量删除文件/目录
func RunRemove(familyId int64, paths ...string) {
	if IsFamilyCloud(familyId) {
		delFamilyCloudFiles(familyId, paths...)
	} else {
		delPersonCloudFiles(familyId, paths...)
	}
}

func delFamilyCloudFiles(familyId int64, paths ...string)  {
	activeUser := GetActiveUser()
	infoList, _, delFileInfos := getBatchTaskInfoList(familyId, paths...)
	if infoList == nil || len(*infoList) == 0 {
		fmt.Println("没有有效的文件可删除")
		return
	}

	// create delete files task
	delParam := &cloudpan.BatchTaskParam{
		TypeFlag: cloudpan.BatchTaskTypeDelete,
		TaskInfos: *infoList,
	}

	taskId, err := activeUser.PanClient().AppCreateBatchTask(familyId, delParam)
	if err != nil {
		fmt.Println("无法删除文件，请稍后重试")
		return
	}
	logger.Verboseln("delete file task id: " + taskId)

	// check task
	time.Sleep(time.Duration(200) * time.Millisecond)
	taskRes, err := activeUser.PanClient().AppCheckBatchTask(cloudpan.BatchTaskTypeDelete, taskId)
	if err != nil || taskRes.TaskStatus != cloudpan.BatchTaskStatusOk {
		fmt.Println("无法删除文件，请稍后重试")
		return
	}

	pnt := func() {
		tb := cmdtable.NewTable(os.Stdout)
		tb.SetHeader([]string{"#", "文件/目录"})
		for k := range *delFileInfos {
			tb.Append([]string{strconv.Itoa(k), (*delFileInfos)[k].Path})
		}
		tb.Render()
	}
	if taskRes.TaskStatus == cloudpan.BatchTaskStatusOk {
		fmt.Println("操作成功, 以下文件/目录已删除, 可在云盘文件回收站找回: ")
		pnt()
	}
}

func delPersonCloudFiles(familyId int64, paths ...string)  {
	activeUser := GetActiveUser()
	infoList, _, delFileInfos := getBatchTaskInfoList(familyId, paths...)
	if infoList == nil || len(*infoList) == 0 {
		fmt.Println("没有有效的文件可删除")
		return
	}

	// create delete files task
	delParam := &cloudpan.BatchTaskParam{
		TypeFlag: cloudpan.BatchTaskTypeDelete,
		TaskInfos: *infoList,
	}

	taskId, err := activeUser.PanClient().CreateBatchTask(delParam)
	if err != nil {
		fmt.Println("无法删除文件，请稍后重试")
		return
	}
	logger.Verboseln("delete file task id: " + taskId)

	// check task
	time.Sleep(time.Duration(200) * time.Millisecond)
	taskRes, err := activeUser.PanClient().CheckBatchTask(cloudpan.BatchTaskTypeDelete, taskId)
	if err != nil || taskRes.TaskStatus != cloudpan.BatchTaskStatusOk {
		fmt.Println("无法删除文件，请稍后重试")
		return
	}

	pnt := func() {
		tb := cmdtable.NewTable(os.Stdout)
		tb.SetHeader([]string{"#", "文件/目录"})
		for k := range *delFileInfos {
			tb.Append([]string{strconv.Itoa(k), (*delFileInfos)[k].Path})
		}
		tb.Render()
	}
	if taskRes.TaskStatus == cloudpan.BatchTaskStatusOk {
		fmt.Println("操作成功, 以下文件/目录已删除, 可在云盘文件回收站找回: ")
		pnt()
	}
}

func getBatchTaskInfoList(familyId int64, paths ...string) (*cloudpan.BatchTaskInfoList, *[]string, *[]*cloudpan.AppFileEntity) {
	activeUser := GetActiveUser()
	failedRmPaths := make([]string, 0, len(paths))
	delFileInfos := make([]*cloudpan.AppFileEntity, 0, len(paths))
	infoList := cloudpan.BatchTaskInfoList{}
	for _, p := range paths {
		absolutePath := path.Clean(activeUser.PathJoin(familyId, p))
		fe, err := activeUser.PanClient().AppFileInfoByPath(familyId, absolutePath)
		if err != nil {
			failedRmPaths = append(failedRmPaths, absolutePath)
			continue
		}
		isFolder := 0
		if fe.IsFolder {
			isFolder = 1
		}
		infoItem := &cloudpan.BatchTaskInfo{
			FileId: fe.FileId,
			FileName: fe.FileName,
			IsFolder: isFolder,
			SrcParentId: fe.ParentId,
		}
		infoList = append(infoList, infoItem)
		delFileInfos = append(delFileInfos, fe)
	}
	return &infoList, &failedRmPaths, &delFileInfos
}