
server:
	docker compose up -d --build redis chat-server

bots:
	docker compose up -d --build chat-bot-1 chat-bot-2
