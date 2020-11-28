package order

import (
	"company/bab/data"
	"company/bab/module/order"
	"company/bab/util/objective"
	"company/bab/view"
	"encoding/json"
	"errors"
	"time"

	"github.com/asaskevich/govalidator"
	"gopkg.in/mgo.v2/bson"
)

var (
	ErrorPaymentType   = errors.New("payment type is not valid")
	ErrorPaymentParams = errors.New("payment parameter error")
	ErrorID            = errors.New("object id not valid")
)

type CustomerAddressForm struct {
	Address      string  `valid:"required" form:"address" json:"address"`
	State        string  `valid:"required" form:"state" json:"state"`
	City         string  `valid:"required" form:"city" json:"city"`
	Country      string  `valid:"required" form:"country" json:"country"`
	PostalCode   string  `valid:"required" form:"postal_code" json:"postal_code"`
	Lat          float64 `form:"latitude" json:"latitude"`
	Lon          float64 `form:"longitude" json:"longitude"`
	Call         string  `form:"call" json:"call"`
	ReceiverName string  `form:"receiver_name" json:"receiver_name"`
}

func (f *CustomerAddressForm) BindModel() (model order.Address) {
	objective.PopulateData(f, &model)
	return model
}

type OrderForm struct {
	Items        []ItemForm    `json:"items"`
	BillingInfo  order.Address `json:"billing_info"`
	ShippingInfo order.Address `json:"shipping_info"`
	Payment      PaymentForm   `json:"payment"`
	Credit       float32       `json:"credit"`
	Address      string        `json:"address"`
}

func (f *OrderForm) BindModel(m *order.Orders) error {
	for _, item := range f.Items {
		if err := m.AddItem(item.Product, item.Count, item.Variants); err != nil {
			return err
		}
	}
	addressBilling := f.BillingInfo
	addressShipping := f.ShippingInfo
	m.BillingInfo = order.BillingInfo{
		BillingAddress: addressBilling,
	}
	if m.BillingInfo.BillingAddress.Shipping {
		m.ShippingInfo = order.ShippingInfo{
			ShippingAddress: addressShipping,
		}
	}
	switch f.Payment.Type {
	case 2:
		m.OnlinePayment = order.PaymentByOnline{
			Webgate:   f.Payment.Webgate,
			CreatedAt: time.Now(),
		}
	case 3:
		m.CourierPayment = order.PaymentByCoordination{
			Description: f.Payment.Description,
			CreatedAt:   time.Now(),
		}
	default:
		m.CoordinatedPayment = order.PaymentByCoordination{
			Description: f.Payment.Description,
			CreatedAt:   time.Now(),
		}
	}
	return nil
}

func (f *OrderShippingForm) Validate() error {
	if f.SendWay == "" {
		switch f.SendWayType {
		case 2, 4:
			f.SendWay = order.Post
		default:
			f.SendWay = order.Courier
		}
	}
	if _, err := govalidator.ValidateStruct(f); err != nil {
		return err
	}
	return data.LocationValidate(f.State, f.City)
}

// func (f *OrderForm) LoadValidate(g *view.GlobalContext) error {
// 	if err := f.Payment.LoadValidate(g); err != nil {
// 		return err
// 	}
// 	if err := json.Unmarshal(
// 		[]byte(g.FormValue("items")), &f.Items); err != nil {
// 		return err
// 	}
// 	for i := range f.Items {
// 		if err := f.Items[i].Validate(); err != nil {
// 			return err
// 		}
// 	}
// 	if err := json.Unmarshal(
// 		[]byte(g.FormValue("shipping_info")), &f.ShippingInfo); err != nil {
// 		return err
// 	}
// 	return f.ShippingInfo.Validate()
// }

type ItemForm struct {
	Product  string   `form:"id" json:"id" valid:"required"`
	Count    int      `form:"count" json:"count" valid:"required"`
	Variants []string `form:"variant" json:"variant" `
}

func (f *ItemForm) Validate() error {
	if !bson.IsObjectIdHex(f.Product) {
		return ErrorID
	}
	if f.Count == 0 {
		f.Count = 1
	}
	return nil
}

type PaymentForm struct {
	Type         int    `form:"type" json:"type"`
	Description  string `form:"description" json:"description"`
	Webgate      string `form:"webgate" json:"webgate"`
	DiscountCode string `form:"code" json:"code"`
}

func (f *PaymentForm) LoadValidate(g *view.GlobalContext) error {
	if err := f.JsonLoad(g); err != nil {
		return err
	}
	return f.Validate(g)
}

func (f *PaymentForm) JsonLoad(g *view.GlobalContext) error {
	return json.Unmarshal([]byte(g.FormValue("payment")), f)
}

func (f *PaymentForm) Validate(g *view.GlobalContext) error {
	switch f.Type {
	case 2:
		if !g.Store.Payment.OnlinePayment {
			return ErrorPaymentType
		}
		if f.Webgate == "" {
			return ErrorPaymentType
		}

	default:
		if !g.Store.Payment.Coordinatad {
			return ErrorPaymentType
		}

	}
	return nil
}

type OrderBillingForm struct {
	FirstName    string  `form:"firstname" json:"firstname" valid:"required"`
	LastName     string  `form:"lastname" json:"lastname" valid:"required"`
	Email        string  `form:"email" json:"email" valid:"required,email"`
	State        string  `form:"state" json:"state" valid:"required"`
	City         string  `form:"city" json:"city" valid:"required"`
	Country      string  `form:"country" json:"country" valid:"required"`
	Address      string  `form:"address" json:"address" valid:"required"`
	Home         string  `form:"home" json:"home"`
	County       string  `form:"county" json:"county"`
	PostalCode   string  `form:"postal_code" json:"postal_code" valid:"required"`
	Call         string  `form:"phone" json:"phone" valid:"required"`
	SendWayType  int     `form:"send_way" json:"send_way"`
	SendWay      string  `form:"send_way_name" json:"send_way_name" valid:"required"`
	Lat          float64 `form:"latitude" json:"lat"`
	Lon          float64 `form:"longitude" json:"lon"`
	ReceiverName string  `form:"receiver_name" json:"receiver_name"`
	Information  string  `form:"information" json:"information"`
	Shipping     bool    `form:"shipping" json:"shipping"`
}

type OrderShippingForm struct {
	FirstName    string  `form:"firstname" json:"firstname" valid:"required"`
	LastName     string  `form:"lastname" json:"lastname" valid:"required"`
	Email        string  `form:"email" json:"email" valid:"required,email"`
	State        string  `form:"state" json:"state" valid:"required"`
	City         string  `form:"country" json:"country" valid:"required"`
	Country      string  `form:"country" json:"country" valid:"required"`
	Address      string  `form:"address" json:"address" valid:"required"`
	Home         string  `form:"home" json:"home"`
	County       string  `form:"county" json:"county"`
	PostalCode   string  `form:"postal_code" json:"postal_code" valid:"required"`
	Call         string  `form:"phone" json:"phone" valid:"required"`
	SendWayType  int     `form:"send_way" json:"send_way"`
	SendWay      string  `form:"send_way_name" json:"send_way_name" valid:"required"`
	Lat          float64 `form:"latitude" json:"lat"`
	Lon          float64 `form:"longitude" json:"lon"`
	ReceiverName string  `form:"receiver_name" json:"receiver_name"`
	Information  string  `form:"information" json:"information"`
}

type BasketItemForm struct {
	Product  string `form:"product" valid:"required" query:"product"`
	Count    int    `form:"count" query:"count"`
	Delete   bool   `form:"delete" query:"delete"`
	InBasket bool   `form:"in_basket" query:"in_basket"`
	Variants string `form:"variants"`
	Email    string `form:"email" json:"email"`
}

func (f *BasketItemForm) Validate() error {
	if _, err := govalidator.ValidateStruct(f); err != nil {
		return err
	}
	if !bson.IsObjectIdHex(f.Product) {
		return ErrorID
	}
	return nil
}
