package main

import (
	"fmt"
	"time"
	"strconv"
	"bytes"
	"math/rand"
	"encoding/json"
	"net"
	"crypto/sha256"
	"os"
	"io/ioutil"
	"strings"
)



func helperThread(){
	logAdd(MESS_INFO, "helperThread запустился")
	for true {
		saveProfiles()
		swiftCounter()

		time.Sleep(time.Second * WAIT_HELPER_CYCLE)
	}
	logAdd(MESS_INFO, "helperThread закончил работу")
}

func getPid(serial string) string{

	var a uint64 = 1
	for _, f := range serial {
		a = a * uint64(f)
	}

	//todo добавить нули если число меньше трех знаков
	b := a % 999
	for b < 100 {
		b = b * 10
	}
	c := (a / 999) % 999
	for c < 100 {
		c = c * 10
	}
	d := ((a / 999) / 999 ) % 999
	for d < 100 {
		d = d * 10
	}
	e := (((a / 999) / 999 ) / 999 ) % 999
	for e < 100 {
		e = e * 10
	}

	var r string
	r = strconv.Itoa(int(b)) + ":" + strconv.Itoa(int(c)) + ":" + strconv.Itoa(int(d)) + ":" + strconv.Itoa(int(e))

	return r
}

func logAdd(TMessage int, Messages string){
	if options.FDebug && typeLog >= TMessage {

		if logFile == nil {
			logFile, _ = os.Create(LOG_NAME)
		}

		//todo наверное стоит убрать, но пока меашет пинг в логах
		if strings.Contains(Messages, "buff (31): {\"TMessage\":18,\"Messages\":null}") || strings.Contains(Messages, "{18 []}") {
			return
		}

		logFile.Write([]byte(fmt.Sprint(time.Now().Format("02 Jan 2006 15:04:05.000000")) + "\t" + messLogText[TMessage] + ":\t" + Messages + "\n"))

		fmt.Println(fmt.Sprint(time.Now().Format("02 Jan 2006 15:04:05.000000")) + "\t" + messLogText[TMessage] + ":\t" + Messages)
	}

}

func createMessage(TMessage int, Messages ...string) Message{
	var mes Message
	mes.TMessage = TMessage
	mes.Messages = Messages
	return mes
}

func randomString(l int) string {
	var result bytes.Buffer
	var temp string
	for i := 0; i < l; {
		if string(randInt(65, 90)) != temp {
			temp = string(randInt(65, 90))
			result.WriteString(temp)
			i++
		}
	}
	return result.String()
}

func randInt(min int, max int) int {
	rand.Seed(time.Now().UTC().UnixNano())
	return min + rand.Intn(max-min)
}

func sendMessageRaw(conn *net.Conn, TMessage int, Messages[] string) bool{
	if conn == nil {
		logAdd(MESS_ERROR, "нет сокета для отправки")
		return false
	}

	var mes Message
	mes.TMessage = TMessage
	mes.Messages = Messages

	out, err := json.Marshal(mes)
	if err == nil && conn != nil {
		_, err = (*conn).Write(out)
		if err == nil {
			return true
		}
	}
	return false
}

func sendMessage(conn *net.Conn, TMessage int, Messages ...string) bool{
	return sendMessageRaw(conn, TMessage, Messages)
}

func getSHA256(str string) string {

	s := sha256.Sum256([]byte(str))
	var r string

	for _, x := range s {
		r = r + fmt.Sprintf("%02x", x)
	}

	return r
}

func delContact(first *Contact, id int) *Contact {
	if first == nil {
		return first
	}

	for first != nil && first.Id == id {
		first = first.Next
	}

	res := first

	for first != nil{
		for first.Next != nil && first.Next.Id == id {
			first.Next = first.Next.Next
		}

		if first.Inner != nil {
			first.Inner = delContact(first.Inner, id)
		}

		first = first.Next
	}

	return res
}

func getContact(first *Contact, id int) *Contact{

	for first != nil {
		if first.Id == id {
			return first
		}

		if first.Inner != nil {
			inner := getContact(first.Inner, id)
			if inner != nil {
				return inner
			}
		}

		first = first.Next
	}

	return nil
}

func getNewId(first *Contact) int {
	if first == nil {
		return 1
	}

	r := 1

	for first != nil {

		if first.Id >= r {
			r = first.Id + 1
		}

		if first.Inner != nil {
			t := getNewId(first.Inner)
			if t >= r {
				r = t + 1
			}
		}

		first = first.Next
	}

	return r
}

func saveProfiles(){
	var list []Profile

	profiles.Range(func(key interface{}, value interface{}) bool{
		list = append(list, *value.(*Profile))
		return true
	})

	b, err := json.Marshal(list)
	if err == nil {
		f, err := os.Create(FILE_PROFILES + ".tmp")
		if err == nil {
			n, err := f.Write(b)
			if n == len(b) && err == nil {
				f.Close()

				os.Remove(FILE_PROFILES)
				os.Rename(FILE_PROFILES + ".tmp", FILE_PROFILES)
			} else {
				f.Close()
				logAdd(MESS_ERROR, "Не удалось сохранить профили: " + fmt.Sprint(err))
			}
		} else {
			logAdd(MESS_ERROR, "Не удалось сохранить профили: " + fmt.Sprint(err))
		}
	} else {
		logAdd(MESS_ERROR, "Не удалось сохранить профили: " + fmt.Sprint(err))
	}
}

func loadProfiles(){
	var list []Profile

	f, err := os.Open(FILE_PROFILES)
	defer f.Close()
	if err == nil {
		b, err := ioutil.ReadAll(f)
		if err == nil {
			err = json.Unmarshal(b, &list)
			if err == nil {
				for _, value := range list {
					profile := value
					profiles.Store(profile.Email, &profile)
				}
			} else {
				logAdd(MESS_ERROR, "Не получилось загрузить профили: " + fmt.Sprint(err))
			}
		} else {
			logAdd(MESS_ERROR, "Не получилось загрузить профили: " + fmt.Sprint(err))
		}
	} else {
		logAdd(MESS_ERROR, "Не получилось загрузить профили: " + fmt.Sprint(err))
	}
}

func saveOptions(){
	b, err := json.Marshal(options)
	if err == nil {
		f, err := os.Create(FILE_OPTIONS + ".tmp")
		if err == nil {
			n, err := f.Write(b)
			if n == len(b) && err == nil {
				f.Close()

				os.Remove(FILE_OPTIONS)
				os.Rename(FILE_OPTIONS + ".tmp", FILE_OPTIONS)
			} else {
				f.Close()
				logAdd(MESS_ERROR, "Не удалось сохранить настройки: " + fmt.Sprint(err))
			}
		} else {
			logAdd(MESS_ERROR, "Не удалось сохранить настройки: " + fmt.Sprint(err))
		}
	} else {
		logAdd(MESS_ERROR, "Не удалось сохранить настройки: " + fmt.Sprint(err))
	}
}

func loadOptions(){
	f, err := os.Open(FILE_OPTIONS)
	defer f.Close()
	if err == nil {
		b, err := ioutil.ReadAll(f)
		if err == nil {
			err = json.Unmarshal(b, &options)
			if err != nil {
				logAdd(MESS_ERROR, "Не получилось загрузить настройки: " + fmt.Sprint(err))
			}
		} else {
			logAdd(MESS_ERROR, "Не получилось загрузить настройки: " + fmt.Sprint(err))
		}
	} else {
		logAdd(MESS_ERROR, "Не получилось загрузить настройки: " + fmt.Sprint(err))
	}
}

//func saveVNCList(){
//
//	b, err := json.Marshal(array_vnc)
//	if err == nil {
//		f, err := os.Create(FILE_VNCLIST + ".tmp")
//		if err == nil {
//			n, err := f.Write(b)
//			if n == len(b) && err == nil {
//				f.Close()
//
//				os.Remove(FILE_VNCLIST)
//				os.Rename(FILE_VNCLIST + ".tmp", FILE_VNCLIST)
//			} else {
//				f.Close()
//				logAdd(MESS_ERROR, "Не удалось сохранить список VNC: " + fmt.Sprint(err))
//			}
//		} else {
//			logAdd(MESS_ERROR, "Не удалось сохранить список VNC: "+fmt.Sprint(err))
//		}
//	} else {
//		logAdd(MESS_ERROR, "Не удалось сохранить список VNC: " + fmt.Sprint(err))
//	}
//}

func loadVNCList(){

	f, err := os.Open(FILE_VNCLIST)
	defer f.Close()
	if err == nil {
		b, err := ioutil.ReadAll(f)
		if err == nil {
			err = json.Unmarshal(b, &arrayVnc)
			if err == nil {
				defaultVnc = 0
				return
			} else {
				logAdd(MESS_ERROR, "Не получилось загрузить список VNC: " + fmt.Sprint(err))
			}
		} else {
			logAdd(MESS_ERROR, "Не получилось загрузить список VNC: " + fmt.Sprint(err))
		}
	} else {
		logAdd(MESS_ERROR, "Не получилось загрузить список VNC: " + fmt.Sprint(err))
	}
}

//пробежимся по профилям, найдем где есть контакты с нашим пид и добавим этот профиль нам
func addClientToProfile(client *Client) {
	profiles.Range(func (key interface {}, value interface {}) bool {
		profile := *value.(*Profile)
		if addClientToContacts(profile.Contacts, client, &profile) {
			//если мы есть хоть в одном конакте этого профиля, пробежимся по ним и отправим свой статус
			profile.clients.Range(func (key interface {}, value interface{}) bool {
				curClient := value.(*Client)
				sendMessage(curClient.Conn, TMESS_STATUS, cleanPid(client.Pid), "1")
				return true
			})
		}
		return true
	})
}

//пробежимся по всем контактам и если есть совпадение, то добавим ссылку на профиль этому клиенту
func addClientToContacts(contact *Contact, client *Client, profile *Profile) bool {
	res := false

	for contact != nil {
		if cleanPid(contact.Pid) == cleanPid(client.Pid) {
			client.profiles.Store(profile.Email, profile)
			res = true
		}

		if contact.Inner != nil {
			innerResult := addClientToContacts(contact.Inner, client, profile)
			if innerResult {
				res = true
			}
		}

		contact = contact.Next
	}

	return res
}

func checkStatuses(curClient *Client, first *Contact) {

	for first != nil {

		if first.Type != "fold" {
			_, exist := clients.Load(cleanPid(first.Pid))
			if exist {
				sendMessage(curClient.Conn, TMESS_STATUS, fmt.Sprint(cleanPid(first.Pid)), "1")
			} else {
				sendMessage(curClient.Conn, TMESS_STATUS, fmt.Sprint(cleanPid(first.Pid)), "0")
			}
		}

		if first.Inner != nil {

			checkStatuses(curClient, first.Inner)
		}

		first = first.Next
	}

}

func getInvisibleEmail(email string) string{

	length := len(email)
	if length > 10 {
		return email[:5] + "*****" + email[length - 5:]
	} else {
		return email[:1] + "*****" + email[length - 1:]
	}
}

func saveCounters() {
	b, err := json.Marshal(counterData)
	if err == nil {
		f, err := os.Create(FILE_COUNTERS)
		if err == nil {
			n, err := f.Write(b)
			if n != len(b) || err != nil {
				logAdd(MESS_ERROR, "Не удалось сохранить счетчики: " + fmt.Sprint(err))
			}
			f.Close()
		} else {
			logAdd(MESS_ERROR, "Не удалось сохранить счетчики: " + fmt.Sprint(err))
		}
	} else {
		logAdd(MESS_ERROR, "Не удалось сохранить счетчики: " + fmt.Sprint(err))
	}
}

func loadCounters(){
	f, err := os.Open(FILE_COUNTERS)
	defer f.Close()
	if err == nil {
		b, err := ioutil.ReadAll(f)
		if err == nil {
			err = json.Unmarshal(b, &counterData)
			if err != nil {
				logAdd(MESS_ERROR, "Не получилось загрузить счетчики: " + fmt.Sprint(err))
			}
		} else {
			logAdd(MESS_ERROR, "Не получилось загрузить счетчики: " + fmt.Sprint(err))
		}
	} else {
		logAdd(MESS_ERROR, "Не получилось загрузить счетчики: " + fmt.Sprint(err))
	}
}

func addCounter(bytes uint64) {
	counterData.mutex.Lock()
	defer counterData.mutex.Unlock()

	counterData.CounterBytes[int(counterData.currentPos.Hour())] = counterData.CounterBytes[int(counterData.currentPos.Hour())] + bytes
	counterData.CounterConnections[int(counterData.currentPos.Hour())] = counterData.CounterConnections[int(counterData.currentPos.Hour())] + 1

	counterData.CounterDayWeekBytes[int(counterData.currentPos.Weekday())] = counterData.CounterDayWeekBytes[int(counterData.currentPos.Weekday())] + bytes
	counterData.CounterDayWeekConnections[int(counterData.currentPos.Weekday())] = counterData.CounterDayWeekConnections[int(counterData.currentPos.Weekday())] + 1

	counterData.CounterDayBytes[int(counterData.currentPos.Day())] = counterData.CounterDayBytes[int(counterData.currentPos.Day())] + bytes
	counterData.CounterDayConnections[int(counterData.currentPos.Day())] = counterData.CounterDayConnections[int(counterData.currentPos.Day())] + 1

	counterData.CounterDayYearBytes[int(counterData.currentPos.YearDay())] = counterData.CounterDayYearBytes[int(counterData.currentPos.YearDay())] + bytes
	counterData.CounterDayYearConnections[int(counterData.currentPos.YearDay())] = counterData.CounterDayYearConnections[int(counterData.currentPos.YearDay())] + 1

	counterData.CounterMonthBytes[int(counterData.currentPos.Month())] = counterData.CounterMonthBytes[int(counterData.currentPos.Month())] + bytes
	counterData.CounterMonthConnections[int(counterData.currentPos.Month())] = counterData.CounterMonthConnections[int(counterData.currentPos.Month())] + 1
}

func swiftCounter() {
	counterData.mutex.Lock()
	defer counterData.mutex.Unlock()

	if time.Now().Hour() != counterData.currentPos.Hour() {
		now := time.Now()
		counterData.CounterBytes[time.Now().Hour()] = 0
		counterData.CounterConnections[time.Now().Hour()] = 0

		if time.Now().Day() != counterData.currentPos.Day(){
			counterData.CounterDayWeekBytes[int(time.Now().Weekday())] = 0
			counterData.CounterDayWeekConnections[int(time.Now().Weekday())] = 0

			counterData.CounterDayBytes[int(time.Now().Day())] = 0
			counterData.CounterDayConnections[int(time.Now().Day())] = 0

			counterData.CounterDayYearBytes[int(time.Now().YearDay())] = 0
			counterData.CounterDayYearConnections[int(time.Now().YearDay())] = 0

			if time.Now().Month() != counterData.currentPos.Month() {
				counterData.CounterMonthBytes[int(time.Now().Month())] = 0
				counterData.CounterMonthConnections[int(time.Now().Month())] = 0
			}
		}

		saveCounters()
		counterData.currentPos = now
	}
}

func cleanPid(pid string) string {
	//todo может потом стоит сюда добавить удаление и других символов
	return strings.Replace(pid, ":", "", -1)
}

func checkError(err error) {
	if err != nil {
		panic(err)
		os.Exit(1)
	}
}