package tan

import (
	"crypto/rand"
	"encoding/base64"
	"log"
	"strings"

	"github.com/Confialink/wallet-accounts/internal/modules/tan/model"
	"golang.org/x/crypto/bcrypt"
)

type HasherVerifier interface {
	// Hash hashes given password
	Hash(password string) (string, error)
	// Verify verifies if given hash and password are match
	Verify(hashed, password string) bool
}

type EventOnSuccessfulUse func(userId, tan string)

type Service struct {
	repository       *Repository
	subscriberRepo   *SubscriberRepository
	hasher           HasherVerifier
	onUseSubscribers []EventOnSuccessfulUse
}

func NewService(repository *Repository, hasher HasherVerifier, subscriberRepo *SubscriberRepository) *Service {
	return &Service{repository: repository, hasher: hasher, subscriberRepo: subscriberRepo}
}

// Add add custom tan to user
func (s *Service) Add(userId, tan string) error {
	hash, err := s.hasher.Hash(tan)
	if nil != err {
		return err
	}
	_, err = s.repository.Create(userId, hash)
	if nil == err {
		s.subscriberRepo.AddSubscriber(userId)
	}
	return err
}

// Verify checks whether tan belongs to user
// foundTan can be used to retrieve found tan, simply pass pointer to tan model pointer
func (s *Service) Verify(userId, tan string, foundTan **model.Tan) bool {
	tans, err := s.repository.FindByUID(userId)
	if nil != err {
		log.Println("tan service failed to verify tan: ", err)
		return false
	}
	for _, tanModel := range tans {
		if s.hasher.Verify(tanModel.Tan, tan) {
			if nil != foundTan {
				*foundTan = tanModel
			}
			return true
		}
	}
	return false
}

// Use verifies if tan is valid and removes it from storage
func (s *Service) Use(userId, tan string) bool {
	var tanModel *model.Tan
	if s.Verify(userId, tan, &tanModel) {
		err := s.repository.Delete(tanModel)
		if nil != err {
			log.Println("tan service failed to delete verified tan: ", err)
			return false
		}
		for _, onUse := range s.onUseSubscribers {
			go onUse(userId, tan)
		}
		return true
	}
	return false
}

func (s *Service) OnSuccessfulUse(onUse EventOnSuccessfulUse) {
	s.onUseSubscribers = append(s.onUseSubscribers, onUse)
}

// Generate generates tan digits string (10 digits length)
func (s *Service) Generate() (string, error) {
	const letters = "0123456789"
	bytes, err := generateRandomBytes(10)
	if err != nil {
		return "", err
	}
	for i, b := range bytes {
		bytes[i] = letters[b%byte(len(letters))]
	}
	return string(bytes), nil
}

// GenerateAndAdd generates specified quantity of tans and assigns them with given user
// it returns generated tans
func (s *Service) GenerateAndAdd(userId string, quantity uint) ([]string, error) {
	var err error
	tans := make([]string, 0, quantity)
	models := make([]*model.Tan, 0, quantity)
	for i := 0; uint(i) < quantity; i++ {
		tan, err := s.Generate()
		if nil != err {
			return nil, err
		}
		hash, err := s.hasher.Hash(tan)
		if nil != err {
			return nil, err
		}
		tans = append(tans, tan)
		models = append(models, &model.Tan{UID: userId, Tan: hash})
	}

	if len(models) > 0 {
		err = s.repository.SaveInTransaction(models)
	}
	if nil == err {
		s.subscriberRepo.AddSubscriber(userId)
	}
	return tans, err
}

// Cancel deactivates previously issued tans
// note that if tan is not passed then all TANs for the given user will be removed
func (s *Service) Cancel(userId string, tans ...string) error {
	if len(tans) > 0 {
		for _, tan := range tans {
			var found *model.Tan
			if s.Verify(userId, tan, &found) {
				err := s.repository.Delete(found)
				if nil != err {
					return err
				}
			} else {
				showNumber := int(float64(len(tan)) * .4)
				maskedTan := tan[:showNumber] + strings.Repeat("*", len(tan)-showNumber)
				log.Printf("[warning] tan service refused to cancel given tan '%s' because the tan has failed verification.", maskedTan)
			}
		}
		return nil
	}
	return s.repository.DeleteByUserId(userId)
}

// Quantity counts number of active tans by user id
func (s *Service) Count(userId string) (uint, error) {
	count, err := s.repository.CountByUserId(userId)
	if nil != err {
		log.Printf("tan service failed to count TANs for user %s: %s", userId, err.Error())
		return 0, err
	}
	return count, nil
}

type BcryptHasherVerifier struct{}

func NewBcryptHasherVerifier() HasherVerifier {
	return &BcryptHasherVerifier{}
}

func (b *BcryptHasherVerifier) Hash(password string) (string, error) {
	result, err := bcrypt.GenerateFromPassword([]byte(password), 8)
	if nil != err {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(result), nil
}

func (b *BcryptHasherVerifier) Verify(hashed, password string) bool {
	hashBytes, err := base64.StdEncoding.DecodeString(hashed)
	if nil != err {
		log.Println("BcryptHasherVerifier.Verify failed: ", err)
		return false
	}
	return nil == bcrypt.CompareHashAndPassword(hashBytes, []byte(password))
}

func generateRandomBytes(n int) ([]byte, error) {
	b := make([]byte, n)
	_, err := rand.Read(b)
	if err != nil {
		return nil, err
	}

	return b, nil
}
