import grpc from 'k6/net/grpc';
import { check, sleep } from 'k6';

const client = new grpc.Client();
client.load(['../proto'], 'telemetry.proto');

// Read certificates
const cert = open('../certs/client.crt');
const key = open('../certs/client.key');
const ca = open('../certs/ca.crt');

export const options = {
    discardResponseBodies: true,
    scenarios: {
        contacts: {
            executor: 'constant-arrival-rate',
            rate: 100,
            timeUnit: '1s',
            duration: '10s',
            preAllocatedVUs: 50,
            maxVUs: 200,
        },
    },
};

export default () => {
    const sinkAddr = __ENV.SINK_ADDR || 'localhost:50051';
    client.connect(sinkAddr, {
        plaintext: false,
        tls: {
            certificate: cert,
            key: key,
            rootCAs: [ca],
        },
    });

    const data = {
        name: 'k6-sensor-mtls',
        value: Math.floor(Math.random() * 100),
        timestamp: new Date().toISOString()
    };

    const response = client.invoke('telemetry.TelemetryService/Publish', data);

    const success = check(response, {
        'status is OK': (r) => r && r.status === grpc.OK,
    });

    if (!success) {
        console.error(`Request failed: status=${response.status} error=${response.error}`);
    }

    client.close();
};
