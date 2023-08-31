package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"strings"
)

func main() {
	// GitLab API 地址
	apiUrl := "http://10.45.211.104/api/v4/groups"

	// 访问令牌
	accessToken := "rPvGLxhK4zGn3obTQP5x"

	// 创建 HTTP 请求
	req, err := http.NewRequest("GET", apiUrl, nil)
	if err != nil {
		fmt.Println(err)
		return
	}

	// 设置访问令牌
	req.Header.Set("PRIVATE-TOKEN", accessToken)

	// 发送 HTTP 请求
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer resp.Body.Close()

	// 解析 JSON 数据
	var groups []map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&groups)
	if err != nil {
		fmt.Println(err)
		return
	}

	// 输出项目信息
	os.Chdir("./repo")
	for _, group := range groups {
		groupName := fmt.Sprintf("%s", group["name"])
		wd, _ := os.Getwd()
		os.MkdirAll(wd+"/"+groupName, 0777)
		// 获取 group 下的项目
		projectsUrl := fmt.Sprintf("%s/%v/projects", apiUrl, group["id"]) + "?per_page=110"
		req, err := http.NewRequest("GET", projectsUrl, nil)
		if err != nil {
			fmt.Println(err)
			continue
		}
		req.Header.Set("PRIVATE-TOKEN", accessToken)
		resp, err := client.Do(req)
		if err != nil {
			fmt.Println(err)
			continue
		}
		defer resp.Body.Close()
		var projects []map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&projects)
		if err != nil {
			fmt.Println(err)
			continue
		}
		os.Chdir("./" + groupName)
		for _, project := range projects {
			projectName := fmt.Sprintf("%v", project["name"])
			httpURL := strings.Replace(fmt.Sprintf("%v", project["http_url_to_repo"]), "gitlab.example.com", "root:meituan123@10.45.211.104", 1)
			if !dirExists("./" + projectName) {
				exec.Command("git", "clone", httpURL).Run()
			}
			os.Chdir("./" + projectName)
			exec.Command("git", "fetch", "-a").Run()
			branches, _ := gitBranches()
			for _, branch := range branches {
				// checkout
				exec.Command("git", "checkout", branch).Run()
			}
			//push到远程：
			newHttpUrl := strings.Replace(fmt.Sprintf("%v", project["http_url_to_repo"]), "gitlab.example.com", "root:meituan123@10.73.255.73", 1)
			err := exec.Command("git", "remote", "set-url", "origin", "--push", newHttpUrl).Run()
			fmt.Println(err)
			err = exec.Command("git", "push", "--all", "origin").Run()
			fmt.Println(err)
			os.Chdir("../")
		}
		os.Chdir("../")
	}
}
func dirExists(path string) bool {
	_, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return false
		} else {
			fmt.Println("获取目录信息时出现错误：", err)
			return false
		}
	} else {
		return true
	}
}

func gitBranches() ([]string, error) {
	cmd := exec.Command("git", "branch", "-r")
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, fmt.Errorf("failed to get stdout pipe: %w", err)
	}

	// 监听 stdout 输出，读取每行分支名称
	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("failed to start cmd: %w", err)
	}
	scanner := bufio.NewScanner(stdout)
	var branches []string
	for scanner.Scan() {
		branch := strings.TrimSpace(strings.ReplaceAll(scanner.Text(), "origin/", ""))
		if len(branch) > 0 {
			branches = append(branches, branch)
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("failed to scan branches: %w", err)
	}

	if err := cmd.Wait(); err != nil {
		return nil, fmt.Errorf("failed to wait cmd: %w", err)
	}

	return branches, nil
}
