package events

type (
	EventArgs struct {
		Sender interface{}
		Args   interface{}
	}
	SubscribeFunc func(args *EventArgs)
)


func NewEventArgs(sender interface{}, args interface{}) *EventArgs {
	return &EventArgs{Sender: sender, Args: args}
}

var	allSubscriptions map[string][]*SubscribeFunc

func Subscribe(event string, subscriber *SubscribeFunc) {
	if allSubscriptions == nil {
		allSubscriptions = make(map[string][]*SubscribeFunc)
	}
	allSubscriptions[event] = append(allSubscriptions[event], subscriber)
}

func UnSubscribe(event string, subscriber *SubscribeFunc) {
	if !(allSubscriptions == nil) {
		tmp := allSubscriptions[event][:0]
		for _, p := range allSubscriptions[event] {
			if subscriber != p {
				tmp = append(tmp, p)
			}
		}
		allSubscriptions[event] = tmp
	}
}

func Publish(event string, args *EventArgs){
	for _, subscriber := range allSubscriptions[event]{
		(*subscriber)(args)
	}
}






