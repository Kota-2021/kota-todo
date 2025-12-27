package repository

import (
	"context"
	"my-portfolio-2025/internal/app/models"
	"time"

	"gorm.io/gorm"
)

// taskRepositoryImpl は TaskRepository インターフェースの具体的な実装です。
type taskRepositoryImpl struct {
	db *gorm.DB // GormのDB接続インスタンス
}

// NewTaskRepository は TaskRepository の新しいインスタンスを作成します。
func NewTaskRepository(db *gorm.DB) TaskRepository {
	return &taskRepositoryImpl{db: db}
}

// Create: 新しいタスクをDBに保存します。
func (r *taskRepositoryImpl) Create(task *models.Task) error {
	// GormのCreateメソッドを呼び出し、データベースにレコードを挿入します。
	result := r.db.Create(task)
	return result.Error
}

// FindAllByUserID: 特定のユーザーIDに紐づく全てのタスクをリストで取得します。
func (r *taskRepositoryImpl) FindAllByUserID(userID uint) ([]models.Task, error) {
	var tasks []models.Task
	// Where条件を使って、UserIDが一致するレコードのみをフィルタリングします。
	result := r.db.Where("user_id = ?", userID).Find(&tasks)

	if result.Error != nil {
		// レコードが見つからない場合もエラーとして扱わず、空のスライスとnilを返すことが多いですが、
		// ここではGormのDBエラーがあれば返します。
		if result.Error == gorm.ErrRecordNotFound {
			return tasks, nil // 見つからない場合は空のリストを返す
		}
		return nil, result.Error
	}
	return tasks, nil
}

// FindByID: IDでタスクを検索します。（認可チェックはService層で行うため、ここでは単純に検索します）
func (r *taskRepositoryImpl) FindByID(taskID uint) (*models.Task, error) {
	var task models.Task
	// Firstは主キーでレコードを検索します。
	result := r.db.First(&task, taskID)

	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, nil // 見つからなかった場合はnilを返す
		}
		return nil, result.Error
	}
	return &task, nil
}

// Update: Taskモデルの変更をDBに保存します。
func (r *taskRepositoryImpl) Update(task *models.Task) error {
	// GormのSaveメソッドは、主キー（ID）に基づいてレコードが存在すれば更新、存在しなければ挿入を行います。
	// 今回はService層で存在チェック済みなので、更新として機能します。
	result := r.db.Save(task)
	return result.Error
}

// Delete: IDを指定してタスクを削除します。
func (r *taskRepositoryImpl) Delete(taskID uint) error {
	// GormのDeleteメソッドを呼び出す。
	// models.Taskがgorm.Modelを含んでいるため、これはデフォルトでソフトデリート（論理削除）になります。
	// 物理削除したい場合は、r.db.Unscoped().Delete(...) を使用する必要がありますが、
	// 通常はソフトデリートが推奨されます。
	result := r.db.Delete(&models.Task{}, taskID)
	return result.Error
}

// FindUpcomingTasks: 指定した日付より前の期限のタスクを取得 (期限切れチェック用)
func (r *taskRepositoryImpl) FindUpcomingTasks(ctx context.Context, threshold time.Time) ([]models.Task, error) {
	var tasks []models.Task
	// 条件:
	// 1. 期限 (due_date) が現在時刻より後
	// 2. 期限 (due_date) が threshold (1時間後など) より前
	// 3. まだ完了していない (status != 'completed')
	// 4. (オプション) すでに通知済みのものは除外するロジックがあればここに追加

	err := r.db.WithContext(ctx).
		Where("due_date > ? AND due_date <= ? AND status != ?", time.Now(), threshold, "completed").
		Find(&tasks).Error

	if err != nil {
		return nil, err
	}
	return tasks, nil
}
