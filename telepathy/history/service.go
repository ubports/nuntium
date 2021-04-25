package history

import (
	"fmt"
	"log"

	"launchpad.net/go-dbus/v1"
)

// HistoryService allows to communicate with message history service through dbus.
type HistoryService struct {
	conn *dbus.Connection
}

func NewHistoryService(conn *dbus.Connection) *HistoryService {
	return &HistoryService{conn}
}

// Returns message identified by parameters from HistoryService.
func (service *HistoryService) GetSingleMessage(accountId, threadId, eventId string) (Message, error) {
	call := dbus.NewMethodCallMessage("com.canonical.HistoryService", "/com/canonical/HistoryService", "com.canonical.HistoryService", "GetSingleEvent")
	eventType := int32(0) // History::EventTypeText
	call.AppendArgs(eventType, accountId, threadId, eventId)
	reply, err := service.conn.SendWithReply(call)
	if err != nil {
		return nil, fmt.Errorf("send with reply error: %w", err)
	}
	if reply.Type == dbus.TypeError {
		return nil, fmt.Errorf("reply is error: %w", reply.AsError())
	}

	msg := map[string]dbus.Variant{}
	if err := reply.Args(&msg); err != nil {
		return nil, fmt.Errorf("reply decoding error: %w", err)
	}

	return Message(msg), nil
}

var ErrorNilHistoryService = fmt.Errorf("nil HistoryService pointer")

// Returns first message identified by eventId from HistoryService.
func (service *HistoryService) GetMessage(eventId string) (Message, error) {
	if service == nil {
		return nil, ErrorNilHistoryService
	}

	// Get event view.
	call := dbus.NewMethodCallMessage("com.canonical.HistoryService", "/com/canonical/HistoryService", "com.canonical.HistoryService", "QueryEvents")
	eventType := int32(0) // History::EventTypeText
	sort := map[string]dbus.Variant(nil)
	filter := map[string]dbus.Variant{
		"filterType":     dbus.Variant{int32(0)}, // FilterTypeStandard
		"filterProperty": dbus.Variant{"eventId"},
		"filterValue":    dbus.Variant{eventId},
		"matchFlags":     dbus.Variant{int32(1)}, // MatchCaseSensitive
	}
	call.AppendArgs(eventType, sort, filter)
	reply, err := service.conn.SendWithReply(call)
	if err != nil {
		return nil, fmt.Errorf("QueryEvents send error: %w", err)
	}
	if reply.Type == dbus.TypeError {
		return nil, fmt.Errorf("QueryEvents reply error: %v", reply.AsError())
	}
	eventView := ""
	if err := reply.Args(&eventView); err != nil {
		return nil, err
	}

	// Destroy event view at end.
	// dbus-send --session --print-reply --dest=com.canonical.HistoryService /com/canonical/HistoryService/eventview2413609620210130164828892 com.canonical.HistoryService.EventView.Destroy
	defer func() {
		destroyCall := dbus.NewMethodCallMessage("com.canonical.HistoryService", dbus.ObjectPath(eventView), "com.canonical.HistoryService.EventView", "Destroy")
		destroyReply, err := service.conn.SendWithReply(destroyCall)
		if err != nil {
			log.Printf("HistoryService.GetMessage: Destroy send error: %v", err)
			return
		}
		if destroyReply.Type == dbus.TypeError {
			log.Printf("HistoryService.GetMessage: Destroy reply error: %v", destroyReply.AsError())
			return
		}
	}()

	// Check if query is valid.
	validCall := dbus.NewMethodCallMessage("com.canonical.HistoryService", dbus.ObjectPath(eventView), "com.canonical.HistoryService.EventView", "IsValid")
	validReply, err := service.conn.SendWithReply(validCall)
	if err != nil {
		return nil, fmt.Errorf("Request validation send error: %w", err)
	}
	if validReply.Type == dbus.TypeError {
		return nil, fmt.Errorf("Request validation reply error: %w", validReply.AsError())
	}
	isValid := false
	if err := validReply.Args(&isValid); err != nil {
		return nil, err
	}
	if !isValid {
		return nil, fmt.Errorf("QueryEvents got invalid query")
	}

	// Get message.
	nextCall := dbus.NewMethodCallMessage("com.canonical.HistoryService", dbus.ObjectPath(eventView), "com.canonical.HistoryService.EventView", "NextPage")
	nextReply, err := service.conn.SendWithReply(nextCall)
	if err != nil {
		return nil, fmt.Errorf("Next page send error: %w", err)
	}
	if nextReply.Type == dbus.TypeError {
		return nil, fmt.Errorf("Next page reply error: %w", nextReply.AsError())
	}
	msgs := []map[string]dbus.Variant(nil)
	if err := nextReply.Args(&msgs); err != nil {
		return nil, fmt.Errorf("Next page reply arguments error: %w", err)
	}
	if len(msgs) > 1 {
		return nil, fmt.Errorf("Too many messages found: %d", len(msgs))
	}
	if len(msgs) == 0 {
		return Message(nil), nil
	}
	return Message(msgs[0]), nil
}
