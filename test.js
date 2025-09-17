import grpc from "k6/net/grpc";
import { check, sleep } from "k6";

const client = new grpc.Client();
client.load(["./proto"], "/products.proto");

// 🔧 Load test configuration
export const options = {
  scenarios: {
    mixed_load: {
      executor: "ramping-vus",
      startVUs: 0,
      stages: [
        { duration: "1m", target: 50 },   // ramp up
        { duration: "2m", target: 50 },   // steady
        { duration: "30s", target: 200 }, // spike
        { duration: "1m", target: 0 },    // ramp down
      ],
      gracefulRampDown: "30s",
    },
  },
  thresholds: {
    grpc_req_duration: ["p(95)<200"], // want 95% < 200ms
    checks: ["rate>0.95"],            // at least 95% OK
  },
};

// ⚖️ Ratio of writes (CreateProduct) vs reads (ListProducts)
const WRITE_RATIO = 0.7; // 70% CreateProduct, 30% ListProducts

export function teardown() {
  client.close();
}

export default function () {
  // Connect once per VU
  if (!client.connected) {
    client.connect("3.16.165.121:50051", { plaintext: true }); // 🔑 Replace with ECS public IP
  }

  if (Math.random() < WRITE_RATIO) {
    // Writers - CreateProduct
    const data = {
      name: `Product-${__VU}-${Date.now()}`,
      description: "Load test product",
      category: "Test Category",
    };

    const response = client.invoke("products.v1.ProductService/CreateProduct", data);
    check(response, {
      "CreateProduct status is OK": (r) => r && r.status === grpc.StatusOK,
    });
  } else {
    // Readers - ListProducts
    const response = client.invoke("products.v1.ProductService/ListProducts", {});
    check(response, {
      "ListProducts status is OK": (r) => r && r.status === grpc.StatusOK,
    });
  }

  sleep(1);
}
