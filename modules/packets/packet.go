package packets

type ServerInitRequestPacket interface {

	Serialize(stream encoding.Stream) error	
}
