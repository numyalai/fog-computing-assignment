package util

func removePacketUDP(arr []PacketUDP, pkt PacketUDP) []PacketUDP {
	for i, r := range arr {

		if r.Id == pkt.Id {
			return append(arr[:i], arr[i+1:]...)
		}
	}
	return arr
}

func removeString(arr []string, id string) []string {
	for i, r := range arr {

		if r == id {
			return append(arr[:i], arr[i+1:]...)
		}
	}
	return arr
}
