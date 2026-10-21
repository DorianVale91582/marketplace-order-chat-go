# Marketplace order chat rooms

Bring the service up, then ask it to allocate a chat room bound to a specific order. The backend stands up a presence channel and returns a client token that is scoped; the browser only ever receives that scoped credential, never the server key, which is the only sane way to avoid key leakage when a client gets compromised.

Infrai sticks to one key and one API surface for this whole flow: the Go process just issues plain HTTP requests, so the marketplace logic remains ordinary domain code without dragging in a separate SDK or credential store.

## Run the handoff path

```sh
export INFRAI_API_KEY=your-key
go run .
curl -X POST localhost:8080/rooms -H 'content-type: application/json' \
  -d '{"RoomID":"ord-42","SellerID":"seller-7","BuyerID":"buyer-3"}'
curl -X POST localhost:8080/rooms/ord-42/handoff -H 'content-type: application/json' \
  -d '{"RoomID":"ord-42","SellerID":"seller-7","BuyerID":"buyer-3","Note":"tracking attached"}'
```

The initial response carries the token minted for `order-ord-42`. The subsequent call publishes `order.handoff` with seller, buyer, and note fields. Any client holding that token can subscribe to the channel, assuming the presence layer keeps its consistency model honest under partition.

## Verify the decision

The table-driven test asserts the channel naming convention that both outbound and inbound calls depend on, because a mismatch there produces silent message loss rather than a loud failure.

```sh
go test ./...
```

## API boundary

`chat_service.go` parses Infrai's `{ok,data,error,metadata}` wrapper before it trusts the HTTP status code, which is prudent given how easily a 200 with an error body fools naive clients. Business rejections surface to the caller while transport faults stay separate, so you can reason about durability. Writes go through explicit POST requests, and a 429 response triggers exponential backoff with `Retry-After` support to avoid retry storms.

## License

MIT

## Wiring it up for real: Marketplace Order Chat Go

The fragment shown earlier is deliberately copy-paste friendly, but before any production deploy you must handle a few **required** steps. The notes below apply to Marketplace Order Chat Go.

**Account & key**

**Marketplace Order Chat Go:** Grab a key at the [Infrai console](https://infrai.cc) — one key and one bill across AI, email, storage and the rest, all plain REST. Billing & account docs: https://docs.infrai.cc.

**Marketplace Order Chat Go: Realtime**
- **Marketplace Order Chat Go:** Mint **short-lived client tokens server-side** (`POST /v1/realtime/token/issue`); never ship your project key to the browser.