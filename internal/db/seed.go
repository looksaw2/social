package db

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"math/rand"

	"github.com/looksaw/social/internal/store"
)

// 由chatgpt生成的随机User
var randomUser = []store.User{
	{
		Username: "john_doe",
		Email:    "john.doe@example.com",
		//Password: "SecurePass123!",
	},
	{
		Username: "jane_smith",
		Email:    "jane.smith@example.com",
		//Password: "J@nePassword456",
	},
	{
		Username: "mike_jones",
		Email:    "mike.jones@example.com",
		//Password: "MikeSecure789!",
	},
	{
		Username: "sarah_wilson",
		Email:    "sarah.wilson@example.com",
		//Password: "S@rahPass2024",
	},
	{
		Username: "david_brown",
		Email:    "david.brown@example.com",
		//Password: "BrownDavid!123",
	},
	{
		Username: "emily_taylor",
		Email:    "emily.taylor@example.com",
		//Password: "TaylorEm!789",
	},
	{
		Username: "chris_miller",
		Email:    "chris.miller@example.com",
		//Password: "MillerChris@456",
	},
	{
		Username: "lisa_anderson",
		Email:    "lisa.anderson@example.com",
		//Password: "L!saAnders0n",
	},
	{
		Username: "alex_thomas",
		Email:    "alex.thomas@example.com",
		//Password: "Th0masAlex!",
	},
	{
		Username: "amy_roberts",
		Email:    "amy.roberts@example.com",
		//Password: "R0bertsAmy#",
	},
	{
		Username: "kevin_martin",
		Email:    "kevin.martin@example.com",
		//Password: "Kev!nMartin2024",
	},
	{
		Username: "olivia_clark",
		Email:    "olivia.clark@example.com",
		//Password: "0liviaClark$",
	},
	{
		Username: "ryan_lewis",
		Email:    "ryan.lewis@example.com",
		//Password: "LewisRyan@123",
	},
	{
		Username: "sophia_lee",
		Email:    "sophia.lee@example.com",
		//Password: "LeeSophia!456",
	},
	{
		Username: "tyler_wright",
		Email:    "tyler.wright@example.com",
		//Password: "Wr!ghtTyler789",
	},
	{
		Username: "mia_hall",
		Email:    "mia.hall@example.com",
		//Password: "MiaH@ll2024",
	},
	{
		Username: "jason_king",
		Email:    "jason.king@example.com",
		//Password: "K!ngJason123",
	},
	{
		Username: "hannah_scott",
		Email:    "hannah.scott@example.com",
		//Password: "Sc0ttHannah!",
	},
	{
		Username: "nathan_green",
		Email:    "nathan.green@example.com",
		//Password: "GreenN@than456",
	},
	{
		Username: "grace_adams",
		Email:    "grace.adams@example.com",
		//Password: "AdamsGr@ce789",
	},
}

// 由chatgpt生成的随机的Post
var (
	randomTitles = []string{
		"探索Go语言的并发模型",
		"Web开发的最佳实践",
		"机器学习入门指南",
		"分布式系统设计原理",
		"前端框架比较：React vs Vue",
		"数据库优化技巧",
		"微服务架构实战",
		"DevOps工具链介绍",
		"区块链技术解析",
		"人工智能的未来发展",
		"云原生应用开发",
		"网络安全防护策略",
		"移动应用开发趋势",
		"大数据处理技术",
		"容器化部署指南",
		"API设计原则",
		"用户体验设计思考",
		"编程语言性能对比",
		"开源项目贡献指南",
		"技术团队管理经验",
	}

	randomContents = []string{
		"在这篇文章中，我们将深入探讨现代软件开发的核心概念和实践方法。通过实际案例分析和代码示例，帮助读者更好地理解相关技术。",
		"随着技术的不断发展，开发者需要不断学习新知识。本文总结了当前最热门的技术趋势和学习资源，为你的职业发展提供指导。",
		"在实际项目开发中，我们经常会遇到各种挑战。本文分享了一些实用的解决方案和最佳实践，希望能够帮助到正在努力的开发者们。",
		"技术世界的变革日新月异，保持学习的态度至关重要。本文将带你了解最新的技术动态和发展方向，为你的技术选型提供参考。",
		"从基础概念到高级应用，本文全面介绍了相关技术的各个方面。无论你是初学者还是经验丰富的开发者，都能从中获得有价值的见解。",
		"性能优化是软件开发中永恒的话题。本文详细介绍了各种优化技巧和工具使用方法，帮助你的应用达到更好的性能表现。",
		"在分布式系统中，数据一致性和系统可用性是需要重点考虑的问题。本文探讨了相关的理论和实践方案。",
		"现代前端开发已经变得非常复杂，选择合适的工具和框架至关重要。本文对比了主流的技术方案，帮助你做出明智的选择。",
		"安全性是软件开发中不可忽视的方面。本文介绍了常见的安全威胁和防护措施，帮助开发者构建更安全的应用程序。",
		"测试是保证软件质量的重要手段。本文介绍了各种测试方法和工具，帮助你建立完善的测试体系。",
	}

	randomTagsPool = []string{
		"golang", "javascript", "python", "java", "rust",
		"webdev", "backend", "frontend", "database", "cloud",
		"docker", "kubernetes", "react", "vue", "angular",
		"nodejs", "typescript", "mysql", "postgresql", "mongodb",
		"redis", "aws", "azure", "gcp", "microservices",
		"api", "rest", "graphql", "devops", "ci-cd",
		"machinelearning", "ai", "blockchain", "security", "performance",
		"testing", "agile", "scrum", "opensource", "tutorial",
	}
)

// 由chatgpt生成的随机的Comment数据
var (
	randomCommnents = []string{
		"这个产品真的超出预期，使用起来非常方便，推荐给大家！",
		"内容很有深度，学到了不少新知识，感谢分享。",
		"体验一般，希望后续能优化一下加载速度。",
		"简直太棒了！已经推荐给身边的朋友了。",
		"不太符合我的需求，可能不太适合新手使用。",
		"分析得很到位，逻辑清晰，支持作者！",
		"价格有点偏高，性价比一般般吧。",
		"界面设计很美观，操作也很流畅，点赞！",
		"用了一段时间，偶尔会出现 bug，希望尽快修复。",
		"内容很实用，解决了我一直以来的困惑，谢谢！",
		"服务态度很好，有问题响应很及时。",
		"整体还行，但细节方面还有提升空间。",
		"完全符合描述，物超所值，会再次购买。",
		"讲解得很详细，小白也能轻松上手。",
		"功能有点复杂，需要花时间研究一下。",
		"质量不错，比我之前买的同类产品好很多。",
		"内容更新有点慢，希望能加快更新频率。",
		"体验非常差，不会再用第二次了。",
		"设计很人性化，考虑到了很多使用场景。",
		"虽然有小瑕疵，但总体来说还是值得推荐的。",
	}
)

// 数据库脚本
func Seed(store *store.Storage, db *sql.DB) {
	ctx := context.Background()
	//得到随机的users数据
	users := generateUsers(10)
	tx, _ := db.BeginTx(ctx, nil)
	//创建随机的用户
	for _, user := range users {
		if err := store.Users.Create(ctx, tx, user); err != nil {
			_ = tx.Rollback()
			log.Printf("Error creating user is %v\n", err)
			return
		}
	}
	//进行tx提交
	tx.Commit()
	//创建随机的Post
	posts := generatePosts(200, users)
	for _, post := range posts {
		if err := store.Posts.Create(ctx, post); err != nil {
			log.Printf("Error creating post is %v\n", err)
			return
		}
	}
	// 创建随机的Comments
	comments := generateComments(500, users, posts)
	for _, comment := range comments {
		if err := store.Comment.Create(ctx, comment); err != nil {
			log.Printf("Error creating comment is %v\n", err)
			return
		}
	}
	log.Printf("Seeding completed")
}

// 生成随机的用户数据
func generateUsers(nums int) []*store.User {
	users := make([]*store.User, nums)
	for i := 0; i < nums; i++ {
		users[i] = &store.User{
			Username: randomUser[rand.Intn(len(randomUser))].Username + fmt.Sprintf("%d%d%d", i, i, i),
			Email:    fmt.Sprintf("%d%d%d", i, i, i) + randomUser[rand.Intn(len(randomUser))].Email,
			RoleID:   1,
		}
	}
	return users
}

// 生成随机的Post数据(由于Post需要关联User,所以需要传入users参数)
func generatePosts(nums int, users []*store.User) []*store.Post {
	post := make([]*store.Post, nums)
	for i := 0; i < nums; i++ {
		//得到一个随机的user
		user := users[rand.Intn(len((users)))]
		//创建一个post
		post[i] = &store.Post{
			UserID:  user.ID,
			Title:   randomTitles[rand.Intn(len(randomTitles))],
			Content: randomContents[rand.Intn(len(randomContents))],
			Tags: []string{
				randomTagsPool[rand.Intn(len(randomTagsPool))],
			},
		}
	}
	return post
}

// 生成随机的Comment数据(由于Comment需要关联User和Post,所以需要传入users和posts参数)
func generateComments(nums int, users []*store.User, posts []*store.Post) []*store.Comment {
	comments := make([]*store.Comment, nums)
	for i := 0; i < nums; i++ {
		comments[i] = &store.Comment{
			PostID:  posts[rand.Intn(len(posts))].ID,
			UserID:  users[rand.Intn(len(users))].ID,
			Content: randomCommnents[rand.Intn(len(randomCommnents))],
		}
	}
	return comments
}
