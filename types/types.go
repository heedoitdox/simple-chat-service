package types

// 소켓 통신에 있어서 사용될 버퍼 크기를 정의한다.
// 만약 이미지나 동영상과 같이 큰 버퍼사이즈를 가지고 있는 데이터를 주고 받게 된다면 키워줘야함. (당연히 소켓 버퍼 사이즈도!)
const (
	SocketBufferSize  = 1024
	MessageBufferSize = 256
)
