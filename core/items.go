package core

import "log"

type Item struct {
	Id         string `json:"id"`
	PriceGross int64  `json:"price_gross"`
	WormType   string `json:"worm_type"`
	WormId     string `json:"worm_id"`
	Status     string `json:"status"`
	CreatedBy  string `json:"created_by"`
}

type InventoryItem struct {
	Id       string `json:"id"`
	WormType string `json:"type"`
	OnMarket bool   `json:"on_market"`
}

func (i *InventoryItem) Type() ItemType {
	if itemType, exist := itemTypes[i.WormType]; exist {
		return itemType
	}

	log.Fatalf("Unknown inventory item type: %s", i.WormType)
	panic("Unknown inventory item type")
}

type ItemType int

const (
	ItemTypeCommon ItemType = iota
	ItemTypeUncommon
	ItemTypeRare
	ItemTypeEpic
)

var itemTypes = map[string]ItemType{
	"common":   ItemTypeCommon,
	"uncommon": ItemTypeUncommon,
	"rare":     ItemTypeRare,
	"epic":     ItemTypeEpic,
}

func (item *Item) Type() ItemType {

	if itemType, exist := itemTypes[item.WormType]; exist {
		return itemType
	}

	log.Fatalf("Unknown item type: %s", item.WormType)
	panic("Unknown item type")
}
