package model

import (
	"errors"
	"strings"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/setting/operation_setting"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

const (
	InviteRebateSourceTopUp        = "topup"
	InviteRebateSourceSubscription = "subscription"
)

type InviteRebate struct {
	Id            int     `json:"id"`
	InviterId     int     `json:"inviter_id" gorm:"index"`
	InviteeId     int     `json:"invitee_id" gorm:"index"`
	SourceType    string  `json:"source_type" gorm:"type:varchar(32);uniqueIndex:idx_invite_rebate_source,priority:1"`
	SourceId      string  `json:"source_id" gorm:"type:varchar(255);uniqueIndex:idx_invite_rebate_source,priority:2"`
	PaymentMethod string  `json:"payment_method" gorm:"type:varchar(50)"`
	BaseQuota     int     `json:"base_quota" gorm:"type:int;default:0"`
	RebateQuota   int     `json:"rebate_quota" gorm:"type:int;default:0"`
	RebateRatio   float64 `json:"rebate_ratio" gorm:"default:0"`
	BaseAmount    float64 `json:"base_amount" gorm:"default:0"`
	Currency      string  `json:"currency" gorm:"type:varchar(8);default:''"`
	CreatedAt     int64   `json:"created_at" gorm:"bigint;index"`
}

type InviteRebateGrantParams struct {
	InviteeId     int
	SourceType    string
	SourceId      string
	PaymentMethod string
	BaseQuota     int
	BaseAmount    float64
	Currency      string
}

type InviteRebateGrantResult struct {
	InviterId   int
	InviteeId   int
	RebateQuota int
}

type InviteRebateView struct {
	Id            int     `json:"id"`
	SourceType    string  `json:"source_type"`
	SourceId      string  `json:"source_id"`
	PaymentMethod string  `json:"payment_method"`
	BaseQuota     int     `json:"base_quota"`
	RebateQuota   int     `json:"rebate_quota"`
	RebateRatio   float64 `json:"rebate_ratio"`
	CreatedAt     int64   `json:"created_at"`
}

func GrantInviteRebate(params InviteRebateGrantParams) (*InviteRebateGrantResult, error) {
	var result *InviteRebateGrantResult
	err := DB.Transaction(func(tx *gorm.DB) error {
		var err error
		result, err = GrantInviteRebateTx(tx, params)
		return err
	})
	if err != nil {
		return nil, err
	}
	return result, nil
}

func GrantInviteRebateTx(tx *gorm.DB, params InviteRebateGrantParams) (*InviteRebateGrantResult, error) {
	if tx == nil {
		return nil, errors.New("tx is nil")
	}
	if !common.InviteRebateEnabled || common.InviteRebateRatio <= 0 || params.BaseQuota <= 0 {
		return nil, nil
	}
	if params.InviteeId <= 0 || params.SourceType == "" || params.SourceId == "" {
		return nil, nil
	}

	var invitee User
	if err := tx.Select("id", "inviter_id").Where("id = ?", params.InviteeId).First(&invitee).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	if invitee.InviterId <= 0 || invitee.InviterId == params.InviteeId {
		return nil, nil
	}

	rebateQuota := int(decimal.NewFromInt(int64(params.BaseQuota)).
		Mul(decimal.NewFromFloat(common.InviteRebateRatio)).
		IntPart())
	if rebateQuota <= 0 {
		return nil, nil
	}

	rebate := &InviteRebate{
		InviterId:     invitee.InviterId,
		InviteeId:     params.InviteeId,
		SourceType:    params.SourceType,
		SourceId:      params.SourceId,
		PaymentMethod: params.PaymentMethod,
		BaseQuota:     params.BaseQuota,
		RebateQuota:   rebateQuota,
		RebateRatio:   common.InviteRebateRatio,
		BaseAmount:    params.BaseAmount,
		Currency:      strings.ToUpper(strings.TrimSpace(params.Currency)),
		CreatedAt:     common.GetTimestamp(),
	}

	result := tx.Clauses(clause.OnConflict{DoNothing: true}).Create(rebate)
	if result.Error != nil {
		return nil, result.Error
	}
	if result.RowsAffected == 0 {
		return nil, nil
	}

	if err := tx.Model(&User{}).Where("id = ?", invitee.InviterId).Updates(map[string]interface{}{
		"aff_quota":   gorm.Expr("aff_quota + ?", rebateQuota),
		"aff_history": gorm.Expr("aff_history + ?", rebateQuota),
	}).Error; err != nil {
		return nil, err
	}

	return &InviteRebateGrantResult{
		InviterId:   invitee.InviterId,
		InviteeId:   params.InviteeId,
		RebateQuota: rebateQuota,
	}, nil
}

func CalculateSubscriptionInviteRebateBaseQuotaTx(tx *gorm.DB, order *SubscriptionOrder, plan *SubscriptionPlan) int {
	if tx == nil || order == nil || plan == nil || order.Money <= 0 {
		return 0
	}

	amount := decimal.NewFromFloat(order.Money)
	quotaPerUnit := decimal.NewFromFloat(common.QuotaPerUnit)
	if subscriptionOrderUsesRechargePrice(order, plan) {
		price := operation_setting.Price
		if price <= 0 {
			price = 1
		}
		topupGroupRatio := 1.0
		if group, err := getUserGroupByIdTx(tx, order.UserId); err == nil {
			if ratio := common.GetTopupGroupRatio(group); ratio > 0 {
				topupGroupRatio = ratio
			}
		}
		amount = amount.Div(decimal.NewFromFloat(price)).Div(decimal.NewFromFloat(topupGroupRatio))
	}
	return int(amount.Mul(quotaPerUnit).IntPart())
}

func subscriptionOrderUsesRechargePrice(order *SubscriptionOrder, plan *SubscriptionPlan) bool {
	if order == nil || plan == nil {
		return false
	}
	currency := strings.ToUpper(strings.TrimSpace(plan.Currency))
	if currency == "CNY" {
		return true
	}
	switch order.PaymentMethod {
	case PaymentMethodStripe:
		return false
	case PaymentMethodCreem:
		return operation_setting.GetQuotaDisplayType() == operation_setting.QuotaDisplayTypeCNY
	default:
		return true
	}
}

func maskInviteRebateSourceId(sourceId string) string {
	sourceId = strings.TrimSpace(sourceId)
	runes := []rune(sourceId)
	if len(runes) <= 8 {
		return sourceId
	}
	return string(runes[:4]) + "..." + string(runes[len(runes)-4:])
}

func GetUserInviteRebates(userId int, pageInfo *common.PageInfo) (rebates []*InviteRebateView, total int64, err error) {
	query := DB.Model(&InviteRebate{}).Where("inviter_id = ?", userId)
	if err = query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	err = query.Select("id, source_type, source_id, payment_method, base_quota, rebate_quota, rebate_ratio, created_at").
		Order("id desc").
		Limit(pageInfo.GetPageSize()).
		Offset(pageInfo.GetStartIdx()).
		Find(&rebates).Error
	for _, rebate := range rebates {
		rebate.SourceId = maskInviteRebateSourceId(rebate.SourceId)
	}
	return rebates, total, err
}
