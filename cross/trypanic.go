package cross

func TryPanic(err error){
	if(err != nil){
		Log.Error(err.Error())
		panic(err)
	}
}