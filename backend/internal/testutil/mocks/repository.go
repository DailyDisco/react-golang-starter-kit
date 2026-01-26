package mocks

import (
	"context"
	"errors"
	"sync"
	"time"

	"react-golang-starter/internal/models"
	"react-golang-starter/internal/repository"
)

// Common errors for mocks
var (
	ErrNotFound = errors.New("not found")
)

// MockSessionRepository implements repository.SessionRepository for testing.
type MockSessionRepository struct {
	mu       sync.RWMutex
	sessions map[uint][]models.UserSession // userID -> sessions
	nextID   uint

	// Error injection
	CreateErr           error
	FindByUserIDErr     error
	DeleteByIDErr       error
	DeleteByUserIDErr   error
	DeleteByTokenErr    error
	UpdateLastActiveErr error
	DeleteExpiredErr    error

	// Call tracking
	CreateCalls           int
	FindByUserIDCalls     int
	DeleteByIDCalls       int
	DeleteByUserIDCalls   int
	DeleteByTokenCalls    int
	UpdateLastActiveCalls int
	DeleteExpiredCalls    int
}

// NewMockSessionRepository creates a new mock session repository.
func NewMockSessionRepository() *MockSessionRepository {
	return &MockSessionRepository{
		sessions: make(map[uint][]models.UserSession),
		nextID:   1,
	}
}

// Create creates a new session record.
func (m *MockSessionRepository) Create(ctx context.Context, session *models.UserSession) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.CreateCalls++
	if m.CreateErr != nil {
		return m.CreateErr
	}

	session.ID = m.nextID
	m.nextID++

	m.sessions[session.UserID] = append(m.sessions[session.UserID], *session)
	return nil
}

// FindByUserID returns all sessions for a user that haven't expired.
func (m *MockSessionRepository) FindByUserID(ctx context.Context, userID uint, now time.Time) ([]models.UserSession, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.FindByUserIDCalls++
	if m.FindByUserIDErr != nil {
		return nil, m.FindByUserIDErr
	}

	var result []models.UserSession
	for _, s := range m.sessions[userID] {
		if s.ExpiresAt.After(now) {
			result = append(result, s)
		}
	}
	return result, nil
}

// DeleteByID deletes a session by ID and user ID.
func (m *MockSessionRepository) DeleteByID(ctx context.Context, sessionID, userID uint) (int64, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.DeleteByIDCalls++
	if m.DeleteByIDErr != nil {
		return 0, m.DeleteByIDErr
	}

	sessions := m.sessions[userID]
	for i, s := range sessions {
		if s.ID == sessionID {
			m.sessions[userID] = append(sessions[:i], sessions[i+1:]...)
			return 1, nil
		}
	}
	return 0, nil
}

// DeleteByUserID deletes all sessions for a user, optionally excluding a token hash.
func (m *MockSessionRepository) DeleteByUserID(ctx context.Context, userID uint, exceptTokenHash string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.DeleteByUserIDCalls++
	if m.DeleteByUserIDErr != nil {
		return m.DeleteByUserIDErr
	}

	if exceptTokenHash == "" {
		delete(m.sessions, userID)
		return nil
	}

	var remaining []models.UserSession
	for _, s := range m.sessions[userID] {
		if s.SessionTokenHash == exceptTokenHash {
			remaining = append(remaining, s)
		}
	}
	m.sessions[userID] = remaining
	return nil
}

// DeleteByTokenHash deletes a session by its token hash.
func (m *MockSessionRepository) DeleteByTokenHash(ctx context.Context, tokenHash string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.DeleteByTokenCalls++
	if m.DeleteByTokenErr != nil {
		return m.DeleteByTokenErr
	}

	for userID, sessions := range m.sessions {
		for i, s := range sessions {
			if s.SessionTokenHash == tokenHash {
				m.sessions[userID] = append(sessions[:i], sessions[i+1:]...)
				return nil
			}
		}
	}
	return nil
}

// UpdateLastActive updates the last_active_at timestamp for a session.
func (m *MockSessionRepository) UpdateLastActive(ctx context.Context, tokenHash string, lastActive time.Time) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.UpdateLastActiveCalls++
	if m.UpdateLastActiveErr != nil {
		return m.UpdateLastActiveErr
	}

	for userID, sessions := range m.sessions {
		for i, s := range sessions {
			if s.SessionTokenHash == tokenHash {
				m.sessions[userID][i].LastActiveAt = lastActive
				return nil
			}
		}
	}
	return nil
}

// DeleteExpired removes all sessions that have expired before the given time.
func (m *MockSessionRepository) DeleteExpired(ctx context.Context, before time.Time) (int64, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.DeleteExpiredCalls++
	if m.DeleteExpiredErr != nil {
		return 0, m.DeleteExpiredErr
	}

	var deleted int64
	for userID, sessions := range m.sessions {
		var remaining []models.UserSession
		for _, s := range sessions {
			if s.ExpiresAt.Before(before) {
				deleted++
			} else {
				remaining = append(remaining, s)
			}
		}
		m.sessions[userID] = remaining
	}
	return deleted, nil
}

// Reset clears all data and resets call counts.
func (m *MockSessionRepository) Reset() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.sessions = make(map[uint][]models.UserSession)
	m.nextID = 1
	m.CreateErr = nil
	m.FindByUserIDErr = nil
	m.DeleteByIDErr = nil
	m.DeleteByUserIDErr = nil
	m.DeleteByTokenErr = nil
	m.UpdateLastActiveErr = nil
	m.DeleteExpiredErr = nil
	m.CreateCalls = 0
	m.FindByUserIDCalls = 0
	m.DeleteByIDCalls = 0
	m.DeleteByUserIDCalls = 0
	m.DeleteByTokenCalls = 0
	m.UpdateLastActiveCalls = 0
	m.DeleteExpiredCalls = 0
}

// AddSession adds a session directly for test setup.
func (m *MockSessionRepository) AddSession(session models.UserSession) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if session.ID == 0 {
		session.ID = m.nextID
		m.nextID++
	}
	m.sessions[session.UserID] = append(m.sessions[session.UserID], session)
}

// GetAllSessions returns all sessions for inspection.
func (m *MockSessionRepository) GetAllSessions() map[uint][]models.UserSession {
	m.mu.RLock()
	defer m.mu.RUnlock()

	result := make(map[uint][]models.UserSession)
	for k, v := range m.sessions {
		sessions := make([]models.UserSession, len(v))
		copy(sessions, v)
		result[k] = sessions
	}
	return result
}

// MockLoginHistoryRepository implements repository.LoginHistoryRepository for testing.
type MockLoginHistoryRepository struct {
	mu      sync.RWMutex
	history map[uint][]models.LoginHistory // userID -> history
	nextID  uint

	// Error injection
	CreateErr      error
	FindByUserErr  error
	CountByUserErr error

	// Call tracking
	CreateCalls      int
	FindByUserCalls  int
	CountByUserCalls int
}

// NewMockLoginHistoryRepository creates a new mock login history repository.
func NewMockLoginHistoryRepository() *MockLoginHistoryRepository {
	return &MockLoginHistoryRepository{
		history: make(map[uint][]models.LoginHistory),
		nextID:  1,
	}
}

// Create records a login attempt.
func (m *MockLoginHistoryRepository) Create(ctx context.Context, record *models.LoginHistory) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.CreateCalls++
	if m.CreateErr != nil {
		return m.CreateErr
	}

	record.ID = m.nextID
	m.nextID++

	m.history[record.UserID] = append(m.history[record.UserID], *record)
	return nil
}

// FindByUserID returns login history for a user with pagination.
func (m *MockLoginHistoryRepository) FindByUserID(ctx context.Context, userID uint, limit, offset int) ([]models.LoginHistory, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.FindByUserCalls++
	if m.FindByUserErr != nil {
		return nil, m.FindByUserErr
	}

	all := m.history[userID]
	if offset >= len(all) {
		return []models.LoginHistory{}, nil
	}

	end := offset + limit
	if end > len(all) {
		end = len(all)
	}

	return all[offset:end], nil
}

// CountByUserID returns the total number of login records for a user.
func (m *MockLoginHistoryRepository) CountByUserID(ctx context.Context, userID uint) (int64, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.CountByUserCalls++
	if m.CountByUserErr != nil {
		return 0, m.CountByUserErr
	}

	return int64(len(m.history[userID])), nil
}

// Reset clears all data and resets call counts.
func (m *MockLoginHistoryRepository) Reset() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.history = make(map[uint][]models.LoginHistory)
	m.nextID = 1
	m.CreateErr = nil
	m.FindByUserErr = nil
	m.CountByUserErr = nil
	m.CreateCalls = 0
	m.FindByUserCalls = 0
	m.CountByUserCalls = 0
}

// AddHistory adds a login history record directly for test setup.
func (m *MockLoginHistoryRepository) AddHistory(record models.LoginHistory) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if record.ID == 0 {
		record.ID = m.nextID
		m.nextID++
	}
	m.history[record.UserID] = append(m.history[record.UserID], record)
}

// GetSuccessfulLogins returns count of successful logins for a user.
func (m *MockLoginHistoryRepository) GetSuccessfulLogins(userID uint) int {
	m.mu.RLock()
	defer m.mu.RUnlock()

	count := 0
	for _, h := range m.history[userID] {
		if h.Success {
			count++
		}
	}
	return count
}

// GetFailedLogins returns count of failed logins for a user.
func (m *MockLoginHistoryRepository) GetFailedLogins(userID uint) int {
	m.mu.RLock()
	defer m.mu.RUnlock()

	count := 0
	for _, h := range m.history[userID] {
		if !h.Success {
			count++
		}
	}
	return count
}

// MockOrganizationRepository implements repository.OrganizationRepository for testing.
type MockOrganizationRepository struct {
	mu     sync.RWMutex
	orgs   map[uint]*models.Organization
	slugs  map[string]uint // slug -> orgID
	nextID uint

	// Error injection
	FindBySlugErr                 error
	FindBySlugWithMembersErr      error
	FindByIDErr                   error
	FindByStripeCustomerIDErr     error
	FindByStripeSubscriptionIDErr error
	CountBySlugErr                error
	CreateErr                     error
	UpdateErr                     error
	UpdatePlanErr                 error
	UpdateStripeCustomerErr       error
	DeleteErr                     error

	// Call tracking
	FindBySlugCalls                 int
	FindBySlugWithMembersCalls      int
	FindByIDCalls                   int
	FindByStripeCustomerIDCalls     int
	FindByStripeSubscriptionIDCalls int
	CountBySlugCalls                int
	CreateCalls                     int
	UpdateCalls                     int
	UpdatePlanCalls                 int
	UpdateStripeCustomerCalls       int
	DeleteCalls                     int
}

// NewMockOrganizationRepository creates a new mock organization repository.
func NewMockOrganizationRepository() *MockOrganizationRepository {
	return &MockOrganizationRepository{
		orgs:   make(map[uint]*models.Organization),
		slugs:  make(map[string]uint),
		nextID: 1,
	}
}

func (m *MockOrganizationRepository) FindBySlug(ctx context.Context, slug string) (*models.Organization, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	m.FindBySlugCalls++
	if m.FindBySlugErr != nil {
		return nil, m.FindBySlugErr
	}
	if id, ok := m.slugs[slug]; ok {
		if org, ok := m.orgs[id]; ok {
			return org, nil
		}
	}
	return nil, ErrNotFound
}

func (m *MockOrganizationRepository) FindBySlugWithMembers(ctx context.Context, slug string) (*models.Organization, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	m.FindBySlugWithMembersCalls++
	if m.FindBySlugWithMembersErr != nil {
		return nil, m.FindBySlugWithMembersErr
	}
	if id, ok := m.slugs[slug]; ok {
		if org, ok := m.orgs[id]; ok {
			return org, nil
		}
	}
	return nil, ErrNotFound
}

func (m *MockOrganizationRepository) FindByID(ctx context.Context, id uint) (*models.Organization, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	m.FindByIDCalls++
	if m.FindByIDErr != nil {
		return nil, m.FindByIDErr
	}
	if org, ok := m.orgs[id]; ok {
		return org, nil
	}
	return nil, ErrNotFound
}

func (m *MockOrganizationRepository) FindByStripeCustomerID(ctx context.Context, customerID string) (*models.Organization, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	m.FindByStripeCustomerIDCalls++
	if m.FindByStripeCustomerIDErr != nil {
		return nil, m.FindByStripeCustomerIDErr
	}
	for _, org := range m.orgs {
		if org.StripeCustomerID != nil && *org.StripeCustomerID == customerID {
			return org, nil
		}
	}
	return nil, ErrNotFound
}

func (m *MockOrganizationRepository) FindByStripeSubscriptionID(ctx context.Context, subID string) (*models.Organization, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	m.FindByStripeSubscriptionIDCalls++
	if m.FindByStripeSubscriptionIDErr != nil {
		return nil, m.FindByStripeSubscriptionIDErr
	}
	for _, org := range m.orgs {
		if org.StripeSubscriptionID != nil && *org.StripeSubscriptionID == subID {
			return org, nil
		}
	}
	return nil, ErrNotFound
}

func (m *MockOrganizationRepository) CountBySlug(ctx context.Context, slug string) (int64, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	m.CountBySlugCalls++
	if m.CountBySlugErr != nil {
		return 0, m.CountBySlugErr
	}
	if _, ok := m.slugs[slug]; ok {
		return 1, nil
	}
	return 0, nil
}

func (m *MockOrganizationRepository) Create(ctx context.Context, org *models.Organization) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.CreateCalls++
	if m.CreateErr != nil {
		return m.CreateErr
	}
	org.ID = m.nextID
	m.nextID++
	orgCopy := *org
	m.orgs[org.ID] = &orgCopy
	m.slugs[org.Slug] = org.ID
	return nil
}

func (m *MockOrganizationRepository) Update(ctx context.Context, org *models.Organization) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.UpdateCalls++
	if m.UpdateErr != nil {
		return m.UpdateErr
	}
	if _, ok := m.orgs[org.ID]; !ok {
		return ErrNotFound
	}
	// Remove old slug mapping if changed
	for slug, id := range m.slugs {
		if id == org.ID && slug != org.Slug {
			delete(m.slugs, slug)
			break
		}
	}
	orgCopy := *org
	m.orgs[org.ID] = &orgCopy
	m.slugs[org.Slug] = org.ID
	return nil
}

func (m *MockOrganizationRepository) UpdatePlan(ctx context.Context, orgID uint, plan models.OrganizationPlan, stripeSubID *string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.UpdatePlanCalls++
	if m.UpdatePlanErr != nil {
		return m.UpdatePlanErr
	}
	org, ok := m.orgs[orgID]
	if !ok {
		return ErrNotFound
	}
	org.Plan = plan
	if stripeSubID != nil {
		org.StripeSubscriptionID = stripeSubID
	}
	return nil
}

func (m *MockOrganizationRepository) UpdateStripeCustomer(ctx context.Context, orgID uint, customerID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.UpdateStripeCustomerCalls++
	if m.UpdateStripeCustomerErr != nil {
		return m.UpdateStripeCustomerErr
	}
	org, ok := m.orgs[orgID]
	if !ok {
		return ErrNotFound
	}
	org.StripeCustomerID = &customerID
	return nil
}

func (m *MockOrganizationRepository) Delete(ctx context.Context, id uint) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.DeleteCalls++
	if m.DeleteErr != nil {
		return m.DeleteErr
	}
	org, ok := m.orgs[id]
	if !ok {
		return ErrNotFound
	}
	delete(m.slugs, org.Slug)
	delete(m.orgs, id)
	return nil
}

// AddOrganization adds an organization directly for test setup.
func (m *MockOrganizationRepository) AddOrganization(org *models.Organization) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if org.ID == 0 {
		org.ID = m.nextID
		m.nextID++
	}
	orgCopy := *org
	m.orgs[org.ID] = &orgCopy
	m.slugs[org.Slug] = org.ID
}

// Reset clears all data and resets call counts.
func (m *MockOrganizationRepository) Reset() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.orgs = make(map[uint]*models.Organization)
	m.slugs = make(map[string]uint)
	m.nextID = 1
	m.FindBySlugErr = nil
	m.FindBySlugWithMembersErr = nil
	m.FindByIDErr = nil
	m.FindByStripeCustomerIDErr = nil
	m.FindByStripeSubscriptionIDErr = nil
	m.CountBySlugErr = nil
	m.CreateErr = nil
	m.UpdateErr = nil
	m.UpdatePlanErr = nil
	m.UpdateStripeCustomerErr = nil
	m.DeleteErr = nil
	m.FindBySlugCalls = 0
	m.FindBySlugWithMembersCalls = 0
	m.FindByIDCalls = 0
	m.FindByStripeCustomerIDCalls = 0
	m.FindByStripeSubscriptionIDCalls = 0
	m.CountBySlugCalls = 0
	m.CreateCalls = 0
	m.UpdateCalls = 0
	m.UpdatePlanCalls = 0
	m.UpdateStripeCustomerCalls = 0
	m.DeleteCalls = 0
}

// MockOrganizationMemberRepository implements repository.OrganizationMemberRepository for testing.
type MockOrganizationMemberRepository struct {
	mu      sync.RWMutex
	members map[uint][]models.OrganizationMember // orgID -> members
	nextID  uint

	// Error injection
	FindByOrgIDErr               error
	FindByOrgIDAndUserIDErr      error
	FindOrgsByUserIDErr          error
	FindOrgsWithRolesByUserIDErr error
	CountByOrgIDAndRoleErr       error
	CountActiveByOrgIDErr        error
	CreateErr                    error
	UpdateErr                    error
	DeleteErr                    error
	DeleteByOrgIDErr             error

	// Call tracking
	FindByOrgIDCalls               int
	FindByOrgIDAndUserIDCalls      int
	FindOrgsByUserIDCalls          int
	FindOrgsWithRolesByUserIDCalls int
	CountByOrgIDAndRoleCalls       int
	CountActiveByOrgIDCalls        int
	CreateCalls                    int
	UpdateCalls                    int
	DeleteCalls                    int
	DeleteByOrgIDCalls             int

	// For FindOrgsByUserID - stores user's org memberships
	userOrgs map[uint][]models.Organization
}

// NewMockOrganizationMemberRepository creates a new mock member repository.
func NewMockOrganizationMemberRepository() *MockOrganizationMemberRepository {
	return &MockOrganizationMemberRepository{
		members:  make(map[uint][]models.OrganizationMember),
		userOrgs: make(map[uint][]models.Organization),
		nextID:   1,
	}
}

func (m *MockOrganizationMemberRepository) FindByOrgID(ctx context.Context, orgID uint) ([]models.OrganizationMember, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	m.FindByOrgIDCalls++
	if m.FindByOrgIDErr != nil {
		return nil, m.FindByOrgIDErr
	}
	return m.members[orgID], nil
}

func (m *MockOrganizationMemberRepository) FindByOrgIDAndUserID(ctx context.Context, orgID, userID uint) (*models.OrganizationMember, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	m.FindByOrgIDAndUserIDCalls++
	if m.FindByOrgIDAndUserIDErr != nil {
		return nil, m.FindByOrgIDAndUserIDErr
	}
	for _, member := range m.members[orgID] {
		if member.UserID == userID {
			return &member, nil
		}
	}
	return nil, ErrNotFound
}

func (m *MockOrganizationMemberRepository) FindOrgsByUserID(ctx context.Context, userID uint) ([]models.Organization, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	m.FindOrgsByUserIDCalls++
	if m.FindOrgsByUserIDErr != nil {
		return nil, m.FindOrgsByUserIDErr
	}
	return m.userOrgs[userID], nil
}

func (m *MockOrganizationMemberRepository) FindOrgsWithRolesByUserID(ctx context.Context, userID uint) ([]repository.OrgWithRole, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	m.FindOrgsWithRolesByUserIDCalls++
	if m.FindOrgsWithRolesByUserIDErr != nil {
		return nil, m.FindOrgsWithRolesByUserIDErr
	}
	var result []repository.OrgWithRole
	for orgID, members := range m.members {
		for _, member := range members {
			if member.UserID == userID && member.Status == models.MemberStatusActive {
				// Find the org in userOrgs
				for _, org := range m.userOrgs[userID] {
					if org.ID == orgID {
						result = append(result, repository.OrgWithRole{
							Organization: org,
							Role:         member.Role,
						})
						break
					}
				}
			}
		}
	}
	return result, nil
}

func (m *MockOrganizationMemberRepository) CountByOrgIDAndRole(ctx context.Context, orgID uint, role models.OrganizationRole) (int64, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	m.CountByOrgIDAndRoleCalls++
	if m.CountByOrgIDAndRoleErr != nil {
		return 0, m.CountByOrgIDAndRoleErr
	}
	var count int64
	for _, member := range m.members[orgID] {
		if member.Role == role {
			count++
		}
	}
	return count, nil
}

func (m *MockOrganizationMemberRepository) CountActiveByOrgID(ctx context.Context, orgID uint) (int64, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	m.CountActiveByOrgIDCalls++
	if m.CountActiveByOrgIDErr != nil {
		return 0, m.CountActiveByOrgIDErr
	}
	var count int64
	for _, member := range m.members[orgID] {
		if member.Status == models.MemberStatusActive {
			count++
		}
	}
	return count, nil
}

func (m *MockOrganizationMemberRepository) Create(ctx context.Context, member *models.OrganizationMember) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.CreateCalls++
	if m.CreateErr != nil {
		return m.CreateErr
	}
	member.ID = m.nextID
	m.nextID++
	m.members[member.OrganizationID] = append(m.members[member.OrganizationID], *member)
	return nil
}

func (m *MockOrganizationMemberRepository) Update(ctx context.Context, member *models.OrganizationMember) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.UpdateCalls++
	if m.UpdateErr != nil {
		return m.UpdateErr
	}
	members := m.members[member.OrganizationID]
	for i, mem := range members {
		if mem.ID == member.ID {
			m.members[member.OrganizationID][i] = *member
			return nil
		}
	}
	return ErrNotFound
}

func (m *MockOrganizationMemberRepository) Delete(ctx context.Context, member *models.OrganizationMember) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.DeleteCalls++
	if m.DeleteErr != nil {
		return m.DeleteErr
	}
	members := m.members[member.OrganizationID]
	for i, mem := range members {
		if mem.ID == member.ID {
			m.members[member.OrganizationID] = append(members[:i], members[i+1:]...)
			return nil
		}
	}
	return ErrNotFound
}

func (m *MockOrganizationMemberRepository) DeleteByOrgID(ctx context.Context, orgID uint) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.DeleteByOrgIDCalls++
	if m.DeleteByOrgIDErr != nil {
		return m.DeleteByOrgIDErr
	}
	delete(m.members, orgID)
	return nil
}

// AddMember adds a member directly for test setup.
func (m *MockOrganizationMemberRepository) AddMember(member models.OrganizationMember) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if member.ID == 0 {
		member.ID = m.nextID
		m.nextID++
	}
	m.members[member.OrganizationID] = append(m.members[member.OrganizationID], member)
}

// SetUserOrgs sets the organizations for a user (for FindOrgsByUserID).
func (m *MockOrganizationMemberRepository) SetUserOrgs(userID uint, orgs []models.Organization) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.userOrgs[userID] = orgs
}

// SetUserOrgsWithRoles sets the organizations with roles for a user (for FindOrgsWithRolesByUserID).
// This helper sets up both userOrgs and members correctly.
func (m *MockOrganizationMemberRepository) SetUserOrgsWithRoles(userID uint, orgsWithRoles []struct {
	Org  models.Organization
	Role models.OrganizationRole
}) {
	m.mu.Lock()
	defer m.mu.Unlock()

	var orgs []models.Organization
	for _, owr := range orgsWithRoles {
		orgs = append(orgs, owr.Org)
		// Also add membership
		member := models.OrganizationMember{
			ID:             m.nextID,
			OrganizationID: owr.Org.ID,
			UserID:         userID,
			Role:           owr.Role,
			Status:         models.MemberStatusActive,
		}
		m.nextID++
		m.members[owr.Org.ID] = append(m.members[owr.Org.ID], member)
	}
	m.userOrgs[userID] = orgs
}

// Reset clears all data and resets call counts.
func (m *MockOrganizationMemberRepository) Reset() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.members = make(map[uint][]models.OrganizationMember)
	m.userOrgs = make(map[uint][]models.Organization)
	m.nextID = 1
	m.FindByOrgIDErr = nil
	m.FindByOrgIDAndUserIDErr = nil
	m.FindOrgsByUserIDErr = nil
	m.FindOrgsWithRolesByUserIDErr = nil
	m.CountByOrgIDAndRoleErr = nil
	m.CountActiveByOrgIDErr = nil
	m.CreateErr = nil
	m.UpdateErr = nil
	m.DeleteErr = nil
	m.DeleteByOrgIDErr = nil
	m.FindByOrgIDCalls = 0
	m.FindByOrgIDAndUserIDCalls = 0
	m.FindOrgsByUserIDCalls = 0
	m.FindOrgsWithRolesByUserIDCalls = 0
	m.CountByOrgIDAndRoleCalls = 0
	m.CountActiveByOrgIDCalls = 0
	m.CreateCalls = 0
	m.UpdateCalls = 0
	m.DeleteCalls = 0
	m.DeleteByOrgIDCalls = 0
}

// MockOrganizationInvitationRepository implements repository.OrganizationInvitationRepository for testing.
type MockOrganizationInvitationRepository struct {
	mu          sync.RWMutex
	invitations map[uint][]models.OrganizationInvitation  // orgID -> invitations
	tokens      map[string]*models.OrganizationInvitation // token -> invitation
	nextID      uint

	// Error injection
	FindByTokenErr                 error
	FindPendingByOrgIDErr          error
	CountPendingByOrgIDAndEmailErr error
	CountPendingByOrgIDErr         error
	CreateErr                      error
	UpdateErr                      error
	DeleteByIDAndOrgIDErr          error
	DeleteByOrgIDErr               error
	DeleteExpiredErr               error

	// Call tracking
	FindByTokenCalls                 int
	FindPendingByOrgIDCalls          int
	CountPendingByOrgIDAndEmailCalls int
	CountPendingByOrgIDCalls         int
	CreateCalls                      int
	UpdateCalls                      int
	DeleteByIDAndOrgIDCalls          int
	DeleteByOrgIDCalls               int
	DeleteExpiredCalls               int
}

// NewMockOrganizationInvitationRepository creates a new mock invitation repository.
func NewMockOrganizationInvitationRepository() *MockOrganizationInvitationRepository {
	return &MockOrganizationInvitationRepository{
		invitations: make(map[uint][]models.OrganizationInvitation),
		tokens:      make(map[string]*models.OrganizationInvitation),
		nextID:      1,
	}
}

func (m *MockOrganizationInvitationRepository) FindByToken(ctx context.Context, token string) (*models.OrganizationInvitation, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	m.FindByTokenCalls++
	if m.FindByTokenErr != nil {
		return nil, m.FindByTokenErr
	}
	if inv, ok := m.tokens[token]; ok {
		return inv, nil
	}
	return nil, ErrNotFound
}

func (m *MockOrganizationInvitationRepository) FindPendingByOrgID(ctx context.Context, orgID uint, now time.Time) ([]models.OrganizationInvitation, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	m.FindPendingByOrgIDCalls++
	if m.FindPendingByOrgIDErr != nil {
		return nil, m.FindPendingByOrgIDErr
	}
	var result []models.OrganizationInvitation
	for _, inv := range m.invitations[orgID] {
		if inv.AcceptedAt == nil && inv.ExpiresAt.After(now) {
			result = append(result, inv)
		}
	}
	return result, nil
}

func (m *MockOrganizationInvitationRepository) CountPendingByOrgIDAndEmail(ctx context.Context, orgID uint, email string, now time.Time) (int64, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	m.CountPendingByOrgIDAndEmailCalls++
	if m.CountPendingByOrgIDAndEmailErr != nil {
		return 0, m.CountPendingByOrgIDAndEmailErr
	}
	var count int64
	for _, inv := range m.invitations[orgID] {
		if inv.Email == email && inv.AcceptedAt == nil && inv.ExpiresAt.After(now) {
			count++
		}
	}
	return count, nil
}

func (m *MockOrganizationInvitationRepository) CountPendingByOrgID(ctx context.Context, orgID uint, now time.Time) (int64, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	m.CountPendingByOrgIDCalls++
	if m.CountPendingByOrgIDErr != nil {
		return 0, m.CountPendingByOrgIDErr
	}
	var count int64
	for _, inv := range m.invitations[orgID] {
		if inv.AcceptedAt == nil && inv.ExpiresAt.After(now) {
			count++
		}
	}
	return count, nil
}

func (m *MockOrganizationInvitationRepository) Create(ctx context.Context, invitation *models.OrganizationInvitation) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.CreateCalls++
	if m.CreateErr != nil {
		return m.CreateErr
	}
	invitation.ID = m.nextID
	m.nextID++
	invCopy := *invitation
	m.invitations[invitation.OrganizationID] = append(m.invitations[invitation.OrganizationID], invCopy)
	m.tokens[invitation.Token] = &invCopy
	return nil
}

func (m *MockOrganizationInvitationRepository) Update(ctx context.Context, invitation *models.OrganizationInvitation) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.UpdateCalls++
	if m.UpdateErr != nil {
		return m.UpdateErr
	}
	invitations := m.invitations[invitation.OrganizationID]
	for i, inv := range invitations {
		if inv.ID == invitation.ID {
			invCopy := *invitation
			m.invitations[invitation.OrganizationID][i] = invCopy
			m.tokens[invitation.Token] = &invCopy
			return nil
		}
	}
	return ErrNotFound
}

func (m *MockOrganizationInvitationRepository) DeleteByIDAndOrgID(ctx context.Context, id, orgID uint) (int64, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.DeleteByIDAndOrgIDCalls++
	if m.DeleteByIDAndOrgIDErr != nil {
		return 0, m.DeleteByIDAndOrgIDErr
	}
	invitations := m.invitations[orgID]
	for i, inv := range invitations {
		if inv.ID == id && inv.AcceptedAt == nil {
			delete(m.tokens, inv.Token)
			m.invitations[orgID] = append(invitations[:i], invitations[i+1:]...)
			return 1, nil
		}
	}
	return 0, nil
}

func (m *MockOrganizationInvitationRepository) DeleteByOrgID(ctx context.Context, orgID uint) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.DeleteByOrgIDCalls++
	if m.DeleteByOrgIDErr != nil {
		return m.DeleteByOrgIDErr
	}
	for _, inv := range m.invitations[orgID] {
		delete(m.tokens, inv.Token)
	}
	delete(m.invitations, orgID)
	return nil
}

func (m *MockOrganizationInvitationRepository) DeleteExpired(ctx context.Context, now time.Time) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.DeleteExpiredCalls++
	if m.DeleteExpiredErr != nil {
		return m.DeleteExpiredErr
	}
	for orgID, invitations := range m.invitations {
		var remaining []models.OrganizationInvitation
		for _, inv := range invitations {
			if inv.ExpiresAt.After(now) || inv.AcceptedAt != nil {
				remaining = append(remaining, inv)
			} else {
				delete(m.tokens, inv.Token)
			}
		}
		m.invitations[orgID] = remaining
	}
	return nil
}

// AddInvitation adds an invitation directly for test setup.
func (m *MockOrganizationInvitationRepository) AddInvitation(invitation models.OrganizationInvitation) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if invitation.ID == 0 {
		invitation.ID = m.nextID
		m.nextID++
	}
	invCopy := invitation
	m.invitations[invitation.OrganizationID] = append(m.invitations[invitation.OrganizationID], invCopy)
	m.tokens[invitation.Token] = &invCopy
}

// Reset clears all data and resets call counts.
func (m *MockOrganizationInvitationRepository) Reset() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.invitations = make(map[uint][]models.OrganizationInvitation)
	m.tokens = make(map[string]*models.OrganizationInvitation)
	m.nextID = 1
	m.FindByTokenErr = nil
	m.FindPendingByOrgIDErr = nil
	m.CountPendingByOrgIDAndEmailErr = nil
	m.CountPendingByOrgIDErr = nil
	m.CreateErr = nil
	m.UpdateErr = nil
	m.DeleteByIDAndOrgIDErr = nil
	m.DeleteByOrgIDErr = nil
	m.DeleteExpiredErr = nil
	m.FindByTokenCalls = 0
	m.FindPendingByOrgIDCalls = 0
	m.CountPendingByOrgIDAndEmailCalls = 0
	m.CountPendingByOrgIDCalls = 0
	m.CreateCalls = 0
	m.UpdateCalls = 0
	m.DeleteByIDAndOrgIDCalls = 0
	m.DeleteByOrgIDCalls = 0
	m.DeleteExpiredCalls = 0
}

// MockSubscriptionRepository implements repository.SubscriptionRepository for testing.
type MockSubscriptionRepository struct {
	mu     sync.RWMutex
	subs   map[uint]*models.Subscription // orgID -> subscription
	nextID uint

	// Error injection
	FindByOrgIDErr error
	CreateErr      error
	UpdateErr      error

	// Call tracking
	FindByOrgIDCalls int
	CreateCalls      int
	UpdateCalls      int
}

// NewMockSubscriptionRepository creates a new mock subscription repository.
func NewMockSubscriptionRepository() *MockSubscriptionRepository {
	return &MockSubscriptionRepository{
		subs:   make(map[uint]*models.Subscription),
		nextID: 1,
	}
}

func (m *MockSubscriptionRepository) FindByOrgID(ctx context.Context, orgID uint) (*models.Subscription, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	m.FindByOrgIDCalls++
	if m.FindByOrgIDErr != nil {
		return nil, m.FindByOrgIDErr
	}
	if sub, ok := m.subs[orgID]; ok {
		return sub, nil
	}
	return nil, ErrNotFound
}

func (m *MockSubscriptionRepository) Create(ctx context.Context, sub *models.Subscription) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.CreateCalls++
	if m.CreateErr != nil {
		return m.CreateErr
	}
	sub.ID = m.nextID
	m.nextID++
	subCopy := *sub
	if sub.OrganizationID != nil {
		m.subs[*sub.OrganizationID] = &subCopy
	}
	return nil
}

func (m *MockSubscriptionRepository) Update(ctx context.Context, sub *models.Subscription) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.UpdateCalls++
	if m.UpdateErr != nil {
		return m.UpdateErr
	}
	subCopy := *sub
	if sub.OrganizationID != nil {
		m.subs[*sub.OrganizationID] = &subCopy
	}
	return nil
}

// AddSubscription adds a subscription directly for test setup.
func (m *MockSubscriptionRepository) AddSubscription(sub models.Subscription) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if sub.ID == 0 {
		sub.ID = m.nextID
		m.nextID++
	}
	subCopy := sub
	if sub.OrganizationID != nil {
		m.subs[*sub.OrganizationID] = &subCopy
	}
}

// Reset clears all data and resets call counts.
func (m *MockSubscriptionRepository) Reset() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.subs = make(map[uint]*models.Subscription)
	m.nextID = 1
	m.FindByOrgIDErr = nil
	m.CreateErr = nil
	m.UpdateErr = nil
	m.FindByOrgIDCalls = 0
	m.CreateCalls = 0
	m.UpdateCalls = 0
}

// MockUserRepository implements repository.UserRepository for testing.
type MockUserRepository struct {
	mu     sync.RWMutex
	users  map[uint]*models.User
	emails map[string]uint
	nextID uint

	// Error injection
	FindByIDErr    error
	FindByEmailErr error
	CreateErr      error
	UpdateErr      error
	DeleteErr      error

	// Call tracking
	FindByIDCalls    int
	FindByEmailCalls int
	CreateCalls      int
	UpdateCalls      int
	DeleteCalls      int
}

// NewMockUserRepository creates a new mock user repository.
func NewMockUserRepository() *MockUserRepository {
	return &MockUserRepository{
		users:  make(map[uint]*models.User),
		emails: make(map[string]uint),
		nextID: 1,
	}
}

func (m *MockUserRepository) FindByID(ctx context.Context, id uint) (*models.User, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	m.FindByIDCalls++
	if m.FindByIDErr != nil {
		return nil, m.FindByIDErr
	}
	if user, ok := m.users[id]; ok {
		return user, nil
	}
	return nil, ErrNotFound
}

func (m *MockUserRepository) FindByEmail(ctx context.Context, email string) (*models.User, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	m.FindByEmailCalls++
	if m.FindByEmailErr != nil {
		return nil, m.FindByEmailErr
	}
	if id, ok := m.emails[email]; ok {
		if user, ok := m.users[id]; ok {
			return user, nil
		}
	}
	return nil, ErrNotFound
}

func (m *MockUserRepository) Create(ctx context.Context, user *models.User) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.CreateCalls++
	if m.CreateErr != nil {
		return m.CreateErr
	}
	user.ID = m.nextID
	m.nextID++
	userCopy := *user
	m.users[user.ID] = &userCopy
	m.emails[user.Email] = user.ID
	return nil
}

func (m *MockUserRepository) Update(ctx context.Context, user *models.User) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.UpdateCalls++
	if m.UpdateErr != nil {
		return m.UpdateErr
	}
	if _, ok := m.users[user.ID]; !ok {
		return ErrNotFound
	}
	// Update email mapping if changed
	for email, id := range m.emails {
		if id == user.ID && email != user.Email {
			delete(m.emails, email)
			break
		}
	}
	userCopy := *user
	m.users[user.ID] = &userCopy
	m.emails[user.Email] = user.ID
	return nil
}

func (m *MockUserRepository) Delete(ctx context.Context, id uint) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.DeleteCalls++
	if m.DeleteErr != nil {
		return m.DeleteErr
	}
	user, ok := m.users[id]
	if !ok {
		return ErrNotFound
	}
	delete(m.emails, user.Email)
	delete(m.users, id)
	return nil
}

// AddUser adds a user directly for test setup.
func (m *MockUserRepository) AddUser(user *models.User) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if user.ID == 0 {
		user.ID = m.nextID
		m.nextID++
	}
	userCopy := *user
	m.users[user.ID] = &userCopy
	m.emails[user.Email] = user.ID
}

// Reset clears all data and resets call counts.
func (m *MockUserRepository) Reset() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.users = make(map[uint]*models.User)
	m.emails = make(map[string]uint)
	m.nextID = 1
	m.FindByIDErr = nil
	m.FindByEmailErr = nil
	m.CreateErr = nil
	m.UpdateErr = nil
	m.DeleteErr = nil
	m.FindByIDCalls = 0
	m.FindByEmailCalls = 0
	m.CreateCalls = 0
	m.UpdateCalls = 0
	m.DeleteCalls = 0
}

// MockSystemSettingRepository implements repository.SystemSettingRepository for testing.
type MockSystemSettingRepository struct {
	mu       sync.RWMutex
	settings map[string]*models.SystemSetting

	// Error injection
	FindAllErr        error
	FindByCategoryErr error
	FindByKeyErr      error
	FindByKeysErr     error
	UpdateByKeyErr    error

	// Call tracking
	FindAllCalls        int
	FindByCategoryCalls int
	FindByKeyCalls      int
	FindByKeysCalls     int
	UpdateByKeyCalls    int
}

// NewMockSystemSettingRepository creates a new mock system setting repository.
func NewMockSystemSettingRepository() *MockSystemSettingRepository {
	return &MockSystemSettingRepository{
		settings: make(map[string]*models.SystemSetting),
	}
}

func (m *MockSystemSettingRepository) FindAll(ctx context.Context) ([]models.SystemSetting, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	m.FindAllCalls++
	if m.FindAllErr != nil {
		return nil, m.FindAllErr
	}
	var result []models.SystemSetting
	for _, s := range m.settings {
		result = append(result, *s)
	}
	return result, nil
}

func (m *MockSystemSettingRepository) FindByCategory(ctx context.Context, category string) ([]models.SystemSetting, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	m.FindByCategoryCalls++
	if m.FindByCategoryErr != nil {
		return nil, m.FindByCategoryErr
	}
	var result []models.SystemSetting
	for _, s := range m.settings {
		if s.Category == category {
			result = append(result, *s)
		}
	}
	return result, nil
}

func (m *MockSystemSettingRepository) FindByKey(ctx context.Context, key string) (*models.SystemSetting, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	m.FindByKeyCalls++
	if m.FindByKeyErr != nil {
		return nil, m.FindByKeyErr
	}
	if s, ok := m.settings[key]; ok {
		return s, nil
	}
	return nil, ErrNotFound
}

func (m *MockSystemSettingRepository) FindByKeys(ctx context.Context, keys []string) ([]models.SystemSetting, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	m.FindByKeysCalls++
	if m.FindByKeysErr != nil {
		return nil, m.FindByKeysErr
	}
	var result []models.SystemSetting
	for _, key := range keys {
		if s, ok := m.settings[key]; ok {
			result = append(result, *s)
		}
	}
	return result, nil
}

func (m *MockSystemSettingRepository) UpdateByKey(ctx context.Context, key string, value []byte, updatedAt string) (int64, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.UpdateByKeyCalls++
	if m.UpdateByKeyErr != nil {
		return 0, m.UpdateByKeyErr
	}
	if s, ok := m.settings[key]; ok {
		s.Value = value
		s.UpdatedAt = updatedAt
		return 1, nil
	}
	return 0, nil
}

// AddSetting adds a setting directly for test setup.
func (m *MockSystemSettingRepository) AddSetting(setting models.SystemSetting) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.settings[setting.Key] = &setting
}

// Reset clears all data and resets call counts.
func (m *MockSystemSettingRepository) Reset() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.settings = make(map[string]*models.SystemSetting)
	m.FindAllErr = nil
	m.FindByCategoryErr = nil
	m.FindByKeyErr = nil
	m.FindByKeysErr = nil
	m.UpdateByKeyErr = nil
	m.FindAllCalls = 0
	m.FindByCategoryCalls = 0
	m.FindByKeyCalls = 0
	m.FindByKeysCalls = 0
	m.UpdateByKeyCalls = 0
}

// MockUsageEventRepository implements repository.UsageEventRepository for testing.
type MockUsageEventRepository struct {
	mu     sync.RWMutex
	events []models.UsageEvent
	nextID uint

	// Error injection
	CreateErr error

	// Call tracking
	CreateCalls int
}

// NewMockUsageEventRepository creates a new mock usage event repository.
func NewMockUsageEventRepository() *MockUsageEventRepository {
	return &MockUsageEventRepository{
		events: make([]models.UsageEvent, 0),
		nextID: 1,
	}
}

func (m *MockUsageEventRepository) Create(ctx context.Context, event *models.UsageEvent) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.CreateCalls++
	if m.CreateErr != nil {
		return m.CreateErr
	}
	event.ID = m.nextID
	m.nextID++
	m.events = append(m.events, *event)
	return nil
}

// GetEvents returns all events for inspection.
func (m *MockUsageEventRepository) GetEvents() []models.UsageEvent {
	m.mu.RLock()
	defer m.mu.RUnlock()
	result := make([]models.UsageEvent, len(m.events))
	copy(result, m.events)
	return result
}

// Reset clears all data and resets call counts.
func (m *MockUsageEventRepository) Reset() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.events = make([]models.UsageEvent, 0)
	m.nextID = 1
	m.CreateErr = nil
	m.CreateCalls = 0
}

// MockUsagePeriodRepository implements repository.UsagePeriodRepository for testing.
type MockUsagePeriodRepository struct {
	mu      sync.RWMutex
	periods map[string]*models.UsagePeriod // key: userID-periodStart or orgID-periodStart
	nextID  uint

	// Error injection
	FindByUserAndPeriodErr error
	FindByOrgAndPeriodErr  error
	FindHistoryByUserErr   error
	FindHistoryByOrgErr    error
	CreateErr              error
	UpdateErr              error
	UpsertErr              error

	// Call tracking
	FindByUserAndPeriodCalls int
	FindByOrgAndPeriodCalls  int
	FindHistoryByUserCalls   int
	FindHistoryByOrgCalls    int
	CreateCalls              int
	UpdateCalls              int
	UpsertCalls              int
}

// NewMockUsagePeriodRepository creates a new mock usage period repository.
func NewMockUsagePeriodRepository() *MockUsagePeriodRepository {
	return &MockUsagePeriodRepository{
		periods: make(map[string]*models.UsagePeriod),
		nextID:  1,
	}
}

func (m *MockUsagePeriodRepository) makeKey(userID, orgID *uint, periodStart string) string {
	if userID != nil {
		return "user-" + string(rune(*userID)) + "-" + periodStart
	}
	if orgID != nil {
		return "org-" + string(rune(*orgID)) + "-" + periodStart
	}
	return ""
}

func (m *MockUsagePeriodRepository) FindByUserAndPeriod(ctx context.Context, userID uint, periodStart, periodEnd string) (*models.UsagePeriod, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	m.FindByUserAndPeriodCalls++
	if m.FindByUserAndPeriodErr != nil {
		return nil, m.FindByUserAndPeriodErr
	}
	key := m.makeKey(&userID, nil, periodStart)
	if p, ok := m.periods[key]; ok {
		return p, nil
	}
	return nil, ErrNotFound
}

func (m *MockUsagePeriodRepository) FindByOrgAndPeriod(ctx context.Context, orgID uint, periodStart, periodEnd string) (*models.UsagePeriod, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	m.FindByOrgAndPeriodCalls++
	if m.FindByOrgAndPeriodErr != nil {
		return nil, m.FindByOrgAndPeriodErr
	}
	key := m.makeKey(nil, &orgID, periodStart)
	if p, ok := m.periods[key]; ok {
		return p, nil
	}
	return nil, ErrNotFound
}

func (m *MockUsagePeriodRepository) FindHistoryByUser(ctx context.Context, userID uint, limit int) ([]models.UsagePeriod, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	m.FindHistoryByUserCalls++
	if m.FindHistoryByUserErr != nil {
		return nil, m.FindHistoryByUserErr
	}
	var result []models.UsagePeriod
	for _, p := range m.periods {
		if p.UserID != nil && *p.UserID == userID {
			result = append(result, *p)
		}
	}
	if len(result) > limit {
		result = result[:limit]
	}
	return result, nil
}

func (m *MockUsagePeriodRepository) FindHistoryByOrg(ctx context.Context, orgID uint, limit int) ([]models.UsagePeriod, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	m.FindHistoryByOrgCalls++
	if m.FindHistoryByOrgErr != nil {
		return nil, m.FindHistoryByOrgErr
	}
	var result []models.UsagePeriod
	for _, p := range m.periods {
		if p.OrganizationID != nil && *p.OrganizationID == orgID {
			result = append(result, *p)
		}
	}
	if len(result) > limit {
		result = result[:limit]
	}
	return result, nil
}

func (m *MockUsagePeriodRepository) Create(ctx context.Context, period *models.UsagePeriod) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.CreateCalls++
	if m.CreateErr != nil {
		return m.CreateErr
	}
	period.ID = m.nextID
	m.nextID++
	key := m.makeKey(period.UserID, period.OrganizationID, period.PeriodStart)
	periodCopy := *period
	m.periods[key] = &periodCopy
	return nil
}

func (m *MockUsagePeriodRepository) Update(ctx context.Context, period *models.UsagePeriod, updates map[string]interface{}) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.UpdateCalls++
	if m.UpdateErr != nil {
		return m.UpdateErr
	}
	key := m.makeKey(period.UserID, period.OrganizationID, period.PeriodStart)
	if p, ok := m.periods[key]; ok {
		if v, exists := updates["usage_totals"]; exists {
			p.UsageTotals = v.(string)
		}
		if v, exists := updates["last_aggregated_at"]; exists {
			str := v.(string)
			p.LastAggregatedAt = &str
		}
		if v, exists := updates["updated_at"]; exists {
			p.UpdatedAt = v.(string)
		}
	}
	return nil
}

func (m *MockUsagePeriodRepository) Upsert(ctx context.Context, period *models.UsagePeriod) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.UpsertCalls++
	if m.UpsertErr != nil {
		return m.UpsertErr
	}
	key := m.makeKey(period.UserID, period.OrganizationID, period.PeriodStart)
	if period.ID == 0 {
		period.ID = m.nextID
		m.nextID++
	}
	periodCopy := *period
	m.periods[key] = &periodCopy
	return nil
}

// AddPeriod adds a period directly for test setup.
func (m *MockUsagePeriodRepository) AddPeriod(period models.UsagePeriod) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if period.ID == 0 {
		period.ID = m.nextID
		m.nextID++
	}
	key := m.makeKey(period.UserID, period.OrganizationID, period.PeriodStart)
	periodCopy := period
	m.periods[key] = &periodCopy
}

// Reset clears all data and resets call counts.
func (m *MockUsagePeriodRepository) Reset() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.periods = make(map[string]*models.UsagePeriod)
	m.nextID = 1
	m.FindByUserAndPeriodErr = nil
	m.FindByOrgAndPeriodErr = nil
	m.FindHistoryByUserErr = nil
	m.FindHistoryByOrgErr = nil
	m.CreateErr = nil
	m.UpdateErr = nil
	m.UpsertErr = nil
	m.FindByUserAndPeriodCalls = 0
	m.FindByOrgAndPeriodCalls = 0
	m.FindHistoryByUserCalls = 0
	m.FindHistoryByOrgCalls = 0
	m.CreateCalls = 0
	m.UpdateCalls = 0
	m.UpsertCalls = 0
}

// MockUsageAlertRepository implements repository.UsageAlertRepository for testing.
type MockUsageAlertRepository struct {
	mu     sync.RWMutex
	alerts map[uint]*models.UsageAlert
	nextID uint

	// Error injection
	FindUnacknowledgedByUserErr error
	FindUnacknowledgedByOrgErr  error
	FindOrCreateErr             error
	AcknowledgeErr              error

	// Call tracking
	FindUnacknowledgedByUserCalls int
	FindUnacknowledgedByOrgCalls  int
	FindOrCreateCalls             int
	AcknowledgeCalls              int
}

// NewMockUsageAlertRepository creates a new mock usage alert repository.
func NewMockUsageAlertRepository() *MockUsageAlertRepository {
	return &MockUsageAlertRepository{
		alerts: make(map[uint]*models.UsageAlert),
		nextID: 1,
	}
}

func (m *MockUsageAlertRepository) FindUnacknowledgedByUser(ctx context.Context, userID uint) ([]models.UsageAlert, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	m.FindUnacknowledgedByUserCalls++
	if m.FindUnacknowledgedByUserErr != nil {
		return nil, m.FindUnacknowledgedByUserErr
	}
	var result []models.UsageAlert
	for _, a := range m.alerts {
		if a.UserID != nil && *a.UserID == userID && !a.Acknowledged {
			result = append(result, *a)
		}
	}
	return result, nil
}

func (m *MockUsageAlertRepository) FindUnacknowledgedByOrg(ctx context.Context, orgID uint) ([]models.UsageAlert, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	m.FindUnacknowledgedByOrgCalls++
	if m.FindUnacknowledgedByOrgErr != nil {
		return nil, m.FindUnacknowledgedByOrgErr
	}
	var result []models.UsageAlert
	for _, a := range m.alerts {
		if a.OrganizationID != nil && *a.OrganizationID == orgID && !a.Acknowledged {
			result = append(result, *a)
		}
	}
	return result, nil
}

func (m *MockUsageAlertRepository) FindOrCreate(ctx context.Context, alert *models.UsageAlert) (bool, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.FindOrCreateCalls++
	if m.FindOrCreateErr != nil {
		return false, m.FindOrCreateErr
	}
	// Check if already exists
	for _, a := range m.alerts {
		if a.UserID != nil && alert.UserID != nil && *a.UserID == *alert.UserID &&
			a.AlertType == alert.AlertType && a.UsageType == alert.UsageType &&
			a.PeriodStart == alert.PeriodStart {
			return false, nil // Already exists
		}
	}
	alert.ID = m.nextID
	m.nextID++
	alertCopy := *alert
	m.alerts[alert.ID] = &alertCopy
	return true, nil
}

func (m *MockUsageAlertRepository) Acknowledge(ctx context.Context, alertID uint, acknowledgedBy uint, acknowledgedAt string) (int64, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.AcknowledgeCalls++
	if m.AcknowledgeErr != nil {
		return 0, m.AcknowledgeErr
	}
	if a, ok := m.alerts[alertID]; ok {
		a.Acknowledged = true
		a.AcknowledgedAt = &acknowledgedAt
		a.AcknowledgedBy = &acknowledgedBy
		return 1, nil
	}
	return 0, nil
}

// AddAlert adds an alert directly for test setup.
func (m *MockUsageAlertRepository) AddAlert(alert models.UsageAlert) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if alert.ID == 0 {
		alert.ID = m.nextID
		m.nextID++
	}
	alertCopy := alert
	m.alerts[alert.ID] = &alertCopy
}

// Reset clears all data and resets call counts.
func (m *MockUsageAlertRepository) Reset() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.alerts = make(map[uint]*models.UsageAlert)
	m.nextID = 1
	m.FindUnacknowledgedByUserErr = nil
	m.FindUnacknowledgedByOrgErr = nil
	m.FindOrCreateErr = nil
	m.AcknowledgeErr = nil
	m.FindUnacknowledgedByUserCalls = 0
	m.FindUnacknowledgedByOrgCalls = 0
	m.FindOrCreateCalls = 0
	m.AcknowledgeCalls = 0
}

// MockIPBlocklistRepository implements repository.IPBlocklistRepository for testing.
type MockIPBlocklistRepository struct {
	mu     sync.RWMutex
	blocks map[uint]*models.IPBlocklist
	nextID uint

	// Error injection
	FindActiveErr error
	CreateErr     error
	DeactivateErr error
	IsBlockedErr  error

	// Call tracking
	FindActiveCalls int
	CreateCalls     int
	DeactivateCalls int
	IsBlockedCalls  int
}

// NewMockIPBlocklistRepository creates a new mock IP blocklist repository.
func NewMockIPBlocklistRepository() *MockIPBlocklistRepository {
	return &MockIPBlocklistRepository{
		blocks: make(map[uint]*models.IPBlocklist),
		nextID: 1,
	}
}

func (m *MockIPBlocklistRepository) FindActive(ctx context.Context) ([]models.IPBlocklist, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	m.FindActiveCalls++
	if m.FindActiveErr != nil {
		return nil, m.FindActiveErr
	}
	var result []models.IPBlocklist
	for _, b := range m.blocks {
		if b.IsActive {
			result = append(result, *b)
		}
	}
	return result, nil
}

func (m *MockIPBlocklistRepository) Create(ctx context.Context, block *models.IPBlocklist) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.CreateCalls++
	if m.CreateErr != nil {
		return m.CreateErr
	}
	block.ID = m.nextID
	m.nextID++
	blockCopy := *block
	m.blocks[block.ID] = &blockCopy
	return nil
}

func (m *MockIPBlocklistRepository) Deactivate(ctx context.Context, id uint, updatedAt string) (int64, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.DeactivateCalls++
	if m.DeactivateErr != nil {
		return 0, m.DeactivateErr
	}
	if b, ok := m.blocks[id]; ok {
		b.IsActive = false
		b.UpdatedAt = updatedAt
		return 1, nil
	}
	return 0, nil
}

func (m *MockIPBlocklistRepository) IsBlocked(ctx context.Context, ip string, now string) (bool, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	m.IsBlockedCalls++
	if m.IsBlockedErr != nil {
		return false, m.IsBlockedErr
	}
	for _, b := range m.blocks {
		if b.IsActive && b.IPAddress == ip {
			if b.ExpiresAt == nil || *b.ExpiresAt > now {
				return true, nil
			}
		}
	}
	return false, nil
}

// AddBlock adds a block directly for test setup.
func (m *MockIPBlocklistRepository) AddBlock(block models.IPBlocklist) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if block.ID == 0 {
		block.ID = m.nextID
		m.nextID++
	}
	blockCopy := block
	m.blocks[block.ID] = &blockCopy
}

// Reset clears all data and resets call counts.
func (m *MockIPBlocklistRepository) Reset() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.blocks = make(map[uint]*models.IPBlocklist)
	m.nextID = 1
	m.FindActiveErr = nil
	m.CreateErr = nil
	m.DeactivateErr = nil
	m.IsBlockedErr = nil
	m.FindActiveCalls = 0
	m.CreateCalls = 0
	m.DeactivateCalls = 0
	m.IsBlockedCalls = 0
}

// MockAnnouncementRepository implements repository.AnnouncementRepository for testing.
type MockAnnouncementRepository struct {
	mu            sync.RWMutex
	announcements map[uint]*models.AnnouncementBanner
	nextID        uint

	// Error injection
	FindAllErr          error
	FindByIDErr         error
	FindActiveErr       error
	CreateErr           error
	UpdateErr           error
	DeleteErr           error
	IncrementDismissErr error
	IncrementViewErr    error

	// Call tracking
	FindAllCalls          int
	FindByIDCalls         int
	FindActiveCalls       int
	CreateCalls           int
	UpdateCalls           int
	DeleteCalls           int
	IncrementDismissCalls int
	IncrementViewCalls    int
}

// NewMockAnnouncementRepository creates a new mock announcement repository.
func NewMockAnnouncementRepository() *MockAnnouncementRepository {
	return &MockAnnouncementRepository{
		announcements: make(map[uint]*models.AnnouncementBanner),
		nextID:        1,
	}
}

func (m *MockAnnouncementRepository) FindAll(ctx context.Context) ([]models.AnnouncementBanner, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	m.FindAllCalls++
	if m.FindAllErr != nil {
		return nil, m.FindAllErr
	}
	var result []models.AnnouncementBanner
	for _, a := range m.announcements {
		result = append(result, *a)
	}
	return result, nil
}

func (m *MockAnnouncementRepository) FindByID(ctx context.Context, id uint) (*models.AnnouncementBanner, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	m.FindByIDCalls++
	if m.FindByIDErr != nil {
		return nil, m.FindByIDErr
	}
	if a, ok := m.announcements[id]; ok {
		copy := *a
		return &copy, nil
	}
	return nil, ErrNotFound
}

func (m *MockAnnouncementRepository) FindActive(ctx context.Context, now string) ([]models.AnnouncementBanner, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	m.FindActiveCalls++
	if m.FindActiveErr != nil {
		return nil, m.FindActiveErr
	}
	var result []models.AnnouncementBanner
	for _, a := range m.announcements {
		if a.IsActive {
			// Simplified check - in real impl would check starts_at/ends_at
			result = append(result, *a)
		}
	}
	return result, nil
}

func (m *MockAnnouncementRepository) Create(ctx context.Context, announcement *models.AnnouncementBanner) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.CreateCalls++
	if m.CreateErr != nil {
		return m.CreateErr
	}
	announcement.ID = m.nextID
	m.nextID++
	copy := *announcement
	m.announcements[announcement.ID] = &copy
	return nil
}

func (m *MockAnnouncementRepository) Update(ctx context.Context, id uint, updates map[string]interface{}) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.UpdateCalls++
	if m.UpdateErr != nil {
		return m.UpdateErr
	}
	if a, ok := m.announcements[id]; ok {
		// Apply updates
		if v, exists := updates["title"]; exists {
			a.Title = v.(string)
		}
		if v, exists := updates["message"]; exists {
			a.Message = v.(string)
		}
		if v, exists := updates["is_active"]; exists {
			a.IsActive = v.(bool)
		}
		if v, exists := updates["type"]; exists {
			a.Type = v.(string)
		}
		return nil
	}
	return ErrNotFound
}

func (m *MockAnnouncementRepository) Delete(ctx context.Context, id uint) (int64, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.DeleteCalls++
	if m.DeleteErr != nil {
		return 0, m.DeleteErr
	}
	if _, ok := m.announcements[id]; ok {
		delete(m.announcements, id)
		return 1, nil
	}
	return 0, nil
}

func (m *MockAnnouncementRepository) IncrementDismissCount(ctx context.Context, id uint) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.IncrementDismissCalls++
	if m.IncrementDismissErr != nil {
		return m.IncrementDismissErr
	}
	if a, ok := m.announcements[id]; ok {
		a.DismissCount++
		return nil
	}
	return nil
}

func (m *MockAnnouncementRepository) IncrementViewCount(ctx context.Context, id uint) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.IncrementViewCalls++
	if m.IncrementViewErr != nil {
		return m.IncrementViewErr
	}
	if a, ok := m.announcements[id]; ok {
		a.ViewCount++
		return nil
	}
	return nil
}

// AddAnnouncement adds an announcement directly for test setup.
func (m *MockAnnouncementRepository) AddAnnouncement(announcement models.AnnouncementBanner) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if announcement.ID == 0 {
		announcement.ID = m.nextID
		m.nextID++
	}
	copy := announcement
	m.announcements[announcement.ID] = &copy
}

// Reset clears all data and resets call counts.
func (m *MockAnnouncementRepository) Reset() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.announcements = make(map[uint]*models.AnnouncementBanner)
	m.nextID = 1
	m.FindAllErr = nil
	m.FindByIDErr = nil
	m.FindActiveErr = nil
	m.CreateErr = nil
	m.UpdateErr = nil
	m.DeleteErr = nil
	m.IncrementDismissErr = nil
	m.IncrementViewErr = nil
	m.FindAllCalls = 0
	m.FindByIDCalls = 0
	m.FindActiveCalls = 0
	m.CreateCalls = 0
	m.UpdateCalls = 0
	m.DeleteCalls = 0
	m.IncrementDismissCalls = 0
	m.IncrementViewCalls = 0
}

// MockEmailTemplateRepository implements repository.EmailTemplateRepository for testing.
type MockEmailTemplateRepository struct {
	mu        sync.RWMutex
	templates map[uint]*models.EmailTemplate
	nextID    uint

	// Error injection
	FindAllErr   error
	FindByIDErr  error
	FindByKeyErr error
	UpdateErr    error

	// Call tracking
	FindAllCalls   int
	FindByIDCalls  int
	FindByKeyCalls int
	UpdateCalls    int
}

// NewMockEmailTemplateRepository creates a new mock email template repository.
func NewMockEmailTemplateRepository() *MockEmailTemplateRepository {
	return &MockEmailTemplateRepository{
		templates: make(map[uint]*models.EmailTemplate),
		nextID:    1,
	}
}

func (m *MockEmailTemplateRepository) FindAll(ctx context.Context) ([]models.EmailTemplate, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	m.FindAllCalls++
	if m.FindAllErr != nil {
		return nil, m.FindAllErr
	}
	var result []models.EmailTemplate
	for _, t := range m.templates {
		result = append(result, *t)
	}
	return result, nil
}

func (m *MockEmailTemplateRepository) FindByID(ctx context.Context, id uint) (*models.EmailTemplate, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	m.FindByIDCalls++
	if m.FindByIDErr != nil {
		return nil, m.FindByIDErr
	}
	if t, ok := m.templates[id]; ok {
		copy := *t
		return &copy, nil
	}
	return nil, ErrNotFound
}

func (m *MockEmailTemplateRepository) FindByKey(ctx context.Context, key string) (*models.EmailTemplate, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	m.FindByKeyCalls++
	if m.FindByKeyErr != nil {
		return nil, m.FindByKeyErr
	}
	for _, t := range m.templates {
		if t.Key == key {
			copy := *t
			return &copy, nil
		}
	}
	return nil, ErrNotFound
}

func (m *MockEmailTemplateRepository) Update(ctx context.Context, id uint, updates map[string]interface{}) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.UpdateCalls++
	if m.UpdateErr != nil {
		return m.UpdateErr
	}
	if t, ok := m.templates[id]; ok {
		// Apply updates
		if v, exists := updates["subject"]; exists {
			t.Subject = v.(string)
		}
		if v, exists := updates["body_html"]; exists {
			t.BodyHTML = v.(string)
		}
		if v, exists := updates["body_text"]; exists {
			t.BodyText = v.(string)
		}
		if v, exists := updates["is_active"]; exists {
			t.IsActive = v.(bool)
		}
		return nil
	}
	return ErrNotFound
}

// AddTemplate adds a template directly for test setup.
func (m *MockEmailTemplateRepository) AddTemplate(template models.EmailTemplate) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if template.ID == 0 {
		template.ID = m.nextID
		m.nextID++
	}
	copy := template
	m.templates[template.ID] = &copy
}

// Reset clears all data and resets call counts.
func (m *MockEmailTemplateRepository) Reset() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.templates = make(map[uint]*models.EmailTemplate)
	m.nextID = 1
	m.FindAllErr = nil
	m.FindByIDErr = nil
	m.FindByKeyErr = nil
	m.UpdateErr = nil
	m.FindAllCalls = 0
	m.FindByIDCalls = 0
	m.FindByKeyCalls = 0
	m.UpdateCalls = 0
}
