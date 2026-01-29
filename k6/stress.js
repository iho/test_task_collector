import grpc from 'k6/net/grpc';
import { check, sleep } from 'k6';

const client = new grpc.Client();
client.load(['../proto'], 'telemetry.proto');

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
        plaintext: true,
    });

    const data = {
        name: 'k6-sensor',
        value: Math.floor(Math.random() * 100),
        timestamp: new Date().toISOString()
    };

    const response = client.invoke('telemetry.TelemetryService/Publish', data);

    const success = check(response, {
        'status is OK': (r) => r && r.status === grpc.StatusOK,
    });

    if (!success) {
        console.error(`Request failed: status=${response.status} error=${response.error}`);
    }

    client.close();
};
