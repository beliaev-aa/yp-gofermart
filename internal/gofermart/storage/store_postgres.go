package storage

import (
	"beliaev-aa/yp-gofermart/internal/gofermart/domain"
	gofermartErrors "beliaev-aa/yp-gofermart/internal/gofermart/errors"
	"database/sql"
	"errors"
	"github.com/lib/pq"
	"go.uber.org/zap"
)

type StorePostgres struct {
	db     *sql.DB
	logger *zap.Logger
}

func NewPostgresStorage(dsn string, logger *zap.Logger) (*StorePostgres, error) {
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		logger.Error("Failed to open database connection", zap.Error(err))
		return nil, err
	}

	// Проверка соединения с базой данных
	err = db.Ping()
	if err != nil {
		logger.Error("Failed to ping database", zap.Error(err))
		return nil, err
	}

	store := &StorePostgres{
		db:     db,
		logger: logger,
	}

	// Инициализация схемы базы данных
	err = store.initSchema()
	if err != nil {
		logger.Error("Failed to initialize database schema", zap.Error(err))
		return nil, err
	}

	return store, nil
}

func (s *StorePostgres) Close() error {
	return s.db.Close()
}

// Инициализация схемы базы данных
func (s *StorePostgres) initSchema() error {
	queries := []string{
		`
		CREATE TABLE IF NOT EXISTS users (
			id SERIAL PRIMARY KEY,
			login TEXT UNIQUE NOT NULL,
			password TEXT NOT NULL,
			balance NUMERIC(18,2) DEFAULT 0
		);
		`,
		`
		CREATE TABLE IF NOT EXISTS orders (
			number TEXT PRIMARY KEY,
			user_id INTEGER NOT NULL REFERENCES users(id),
			status TEXT NOT NULL,
			accrual NUMERIC(18,2) DEFAULT 0,
			uploaded_at TIMESTAMP NOT NULL,
		    processing BOOLEAN DEFAULT FALSE
		);
		`,
		`
		CREATE TABLE IF NOT EXISTS withdrawals (
			id SERIAL PRIMARY KEY,
			order_number TEXT NOT NULL,
			user_id INTEGER NOT NULL REFERENCES users(id),
			sum NUMERIC(18,2) NOT NULL,
			processed_at TIMESTAMP NOT NULL
		);
		`,
	}

	for _, query := range queries {
		_, err := s.db.Exec(query)
		if err != nil {
			return err
		}
	}

	return nil
}

// GetUserByLogin Получение пользователя по логину
func (s *StorePostgres) GetUserByLogin(login string) (*domain.User, error) {
	s.logger.Info("Getting user by login", zap.String("login", login))
	var user domain.User
	err := s.db.QueryRow(`
		SELECT id, login, password, balance FROM users WHERE login = $1
	`, login).Scan(&user.ID, &user.Login, &user.Password, &user.Balance)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			s.logger.Warn("User not found", zap.String("login", login))
			return nil, nil
		}
		s.logger.Error("Failed to get user by login", zap.Error(err))
		return nil, err
	}
	s.logger.Info("User retrieved successfully", zap.String("login", login))
	return &user, nil
}

// SaveUser Сохранение нового пользователя
func (s *StorePostgres) SaveUser(user domain.User) error {
	s.logger.Info("Saving new user", zap.String("login", user.Login))
	_, err := s.db.Exec(`
        INSERT INTO users (login, password) VALUES ($1, $2)
    `, user.Login, user.Password)
	if err != nil {
		var pgErr *pq.Error
		if errors.As(err, &pgErr) {
			if pgErr.Code == "23505" { // Код ошибки уникального ограничения
				s.logger.Warn("Login already exists", zap.String("login", user.Login))
				return gofermartErrors.ErrLoginAlreadyExists
			}
		}
		s.logger.Error("Failed to save user", zap.Error(err))
		return err
	}
	s.logger.Info("User saved successfully", zap.String("login", user.Login))
	return nil
}

// UpdateUserBalance Обновление баланса пользователя
func (s *StorePostgres) UpdateUserBalance(userID int, amount float64) error {
	s.logger.Info("Updating user balance", zap.Int("userID", userID), zap.Float64("amount", amount))
	_, err := s.db.Exec(`
		UPDATE users SET balance = balance + $1 WHERE id = $2
	`, amount, userID)
	if err != nil {
		s.logger.Error("Failed to update user balance", zap.Error(err))
		return err
	}
	s.logger.Info("User balance updated successfully", zap.Int("userID", userID))
	return nil
}

// AddWithdrawal Добавление записи о выводе средств
func (s *StorePostgres) AddWithdrawal(withdrawal domain.Withdrawal) error {
	s.logger.Info("Adding withdrawal", zap.Int("userID", withdrawal.UserID), zap.String("order", withdrawal.Order))
	_, err := s.db.Exec(`
		INSERT INTO withdrawals (order_number, user_id, sum, processed_at)
		VALUES ($1, $2, $3, $4)
	`, withdrawal.Order, withdrawal.UserID, withdrawal.Sum, withdrawal.ProcessedAt)
	if err != nil {
		s.logger.Error("Failed to add withdrawal", zap.Error(err))
		return err
	}
	s.logger.Info("Withdrawal added successfully", zap.Int("userID", withdrawal.UserID), zap.String("order", withdrawal.Order))
	return nil
}

// GetWithdrawalsByUserID Получение списка выводов пользователя
func (s *StorePostgres) GetWithdrawalsByUserID(userID int) ([]domain.Withdrawal, error) {
	s.logger.Info("Getting withdrawals for user", zap.Int("userID", userID))
	rows, err := s.db.Query(`
		SELECT id, order_number, user_id, sum, processed_at
		FROM withdrawals WHERE user_id = $1
		ORDER BY processed_at DESC
	`, userID)
	if err != nil {
		s.logger.Error("Failed to get withdrawals", zap.Error(err))
		return nil, err
	}
	defer rows.Close()

	var withdrawals []domain.Withdrawal
	for rows.Next() {
		var w domain.Withdrawal
		err := rows.Scan(&w.ID, &w.Order, &w.UserID, &w.Sum, &w.ProcessedAt)
		if err != nil {
			s.logger.Error("Failed to scan withdrawal", zap.Error(err))
			continue
		}
		withdrawals = append(withdrawals, w)
	}

	if err = rows.Err(); err != nil {
		s.logger.Error("Row iteration error", zap.Error(err))
		return nil, err
	}

	s.logger.Info("Withdrawals retrieved successfully", zap.Int("userID", userID), zap.Int("count", len(withdrawals)))
	return withdrawals, nil
}

// GetOrdersByUserID Получение списка заказов пользователя
func (s *StorePostgres) GetOrdersByUserID(userID int) ([]domain.Order, error) {
	s.logger.Info("Getting orders for user", zap.Int("userID", userID))
	rows, err := s.db.Query(`
		SELECT number, user_id, status, accrual, uploaded_at
		FROM orders WHERE user_id = $1
		ORDER BY uploaded_at DESC
	`, userID)
	if err != nil {
		s.logger.Error("Failed to get orders", zap.Error(err))
		return nil, err
	}
	defer rows.Close()

	var orders []domain.Order
	for rows.Next() {
		var o domain.Order
		err := rows.Scan(&o.Number, &o.UserID, &o.Status, &o.Accrual, &o.UploadedAt)
		if err != nil {
			s.logger.Error("Failed to scan order", zap.Error(err))
			continue
		}
		orders = append(orders, o)
	}

	if err = rows.Err(); err != nil {
		s.logger.Error("Row iteration error", zap.Error(err))
		return nil, err
	}

	s.logger.Info("Orders retrieved successfully", zap.Int("userID", userID), zap.Int("count", len(orders)))
	return orders, nil
}

// AddOrder Добавление нового заказа
func (s *StorePostgres) AddOrder(order domain.Order) error {
	s.logger.Info("Adding new order", zap.String("number", order.Number), zap.Int("userID", order.UserID))
	_, err := s.db.Exec(`
        INSERT INTO orders (number, user_id, status, accrual, uploaded_at)
        VALUES ($1, $2, $3, $4, $5)
    `, order.Number, order.UserID, order.Status, order.Accrual, order.UploadedAt)
	if err != nil {
		var pgErr *pq.Error
		if errors.As(err, &pgErr) {
			if pgErr.Code == "23505" { // Код ошибки уникального ограничения
				s.logger.Warn("Order number already exists", zap.String("number", order.Number))
				return gofermartErrors.ErrOrderAlreadyExists
			}
		}
		s.logger.Error("Failed to add order", zap.Error(err))
		return err
	}
	s.logger.Info("Order added successfully", zap.String("number", order.Number))
	return nil
}

// GetOrderByNumber Получение заказа по номеру
func (s *StorePostgres) GetOrderByNumber(number string) (*domain.Order, error) {
	s.logger.Info("Getting order by number", zap.String("number", number))
	var order domain.Order
	err := s.db.QueryRow(`
		SELECT number, user_id, status, accrual, uploaded_at FROM orders WHERE number = $1
	`, number).Scan(&order.Number, &order.UserID, &order.Status, &order.Accrual, &order.UploadedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			s.logger.Warn("Order not found", zap.String("number", number))
			return nil, nil
		}
		s.logger.Error("Failed to get order by number", zap.Error(err))
		return nil, err
	}
	s.logger.Info("Order retrieved successfully", zap.String("number", number))
	return &order, nil
}

// UpdateOrder Обновление заказа
func (s *StorePostgres) UpdateOrder(order domain.Order) error {
	s.logger.Info("Updating order", zap.String("number", order.Number))
	_, err := s.db.Exec(`
		UPDATE orders SET status = $1, accrual = $2 WHERE number = $3
	`, order.Status, order.Accrual, order.Number)
	if err != nil {
		s.logger.Error("Failed to update order", zap.Error(err))
		return err
	}
	s.logger.Info("Order updated successfully", zap.String("number", order.Number))
	return nil
}

func (s *StorePostgres) GetOrdersForProcessing() ([]domain.Order, error) {
	var orders []domain.Order
	// Выбираем заказы, которые еще не обрабатываются (processing = FALSE)
	query := `
		SELECT number, status, accrual, user_id, processing
		FROM orders
		WHERE status IN ($1, $2, $3) AND processing = FALSE
		FOR UPDATE`
	rows, err := s.db.Query(query, domain.OrderStatusNew, domain.OrderStatusRegistered, domain.OrderStatusProcessing)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var order domain.Order
		err := rows.Scan(&order.Number, &order.Status, &order.Accrual, &order.UserID, &order.Processing)
		if err != nil {
			return nil, err
		}
		orders = append(orders, order)
	}

	return orders, nil
}

func (s *StorePostgres) LockOrderForProcessing(orderNumber string) error {
	query := `
		UPDATE orders 
		SET processing = TRUE 
		WHERE number = $1`
	_, err := s.db.Exec(query, orderNumber)
	return err
}

func (s *StorePostgres) UnlockOrder(orderNumber string) error {
	query := `
		UPDATE orders 
		SET processing = FALSE 
		WHERE number = $1`
	_, err := s.db.Exec(query, orderNumber)
	return err
}

func (s *StorePostgres) GetUserBalance(login string) (balance, withdrawal float64, err error) {
	query := `
		SELECT 
			u.balance, 
			COALESCE(SUM(w.sum), 0) AS total_withdrawals
		FROM 
			users u
		LEFT JOIN 
			withdrawals w ON u.id = w.user_id
		WHERE 
			u.login = $1
		GROUP BY 
			u.balance;
	`

	err = s.db.QueryRow(query, login).Scan(&balance, &withdrawal)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, 0, gofermartErrors.ErrUserNotFound
		}
		return 0, 0, err
	}

	return balance, withdrawal, nil
}
