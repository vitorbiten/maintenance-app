import http from 'k6/http';
import { check, group, sleep } from 'k6';

export const options = {
  stages: [{ target: 50, duration: '30s' }, { target: 200, duration: '30s' }, { target: 1000, duration: '30s' }],
  thresholds: {
    'http_req_duration': ['p(95)<500', 'p(99)<1500'],
    'http_req_duration{name:Create}': ['avg<600', 'max<1000'],
    'http_req_duration{name:Update}': ['avg<600', 'max<1000'],
    'http_req_duration{name:Get}': ['avg<600', 'max<1000'],
    'http_req_duration{name:Delete}': ['avg<600', 'max<1000'],
  },
};

function randomString(length, charset = '') {
  if (!charset) charset = 'abcdefghijklmnopqrstuvwxyz';
  let res = '';
  while (length--) res += charset[(Math.random() * charset.length) | 0];
  return res;
}

const EMAIL = `${randomString(10)}@example.com`; // Set your own email or `${randomString(10)}@example.com`;
const NICKNAME = `${randomString(15)}`; // Set your own email or `${randomString(10)}@example.com`;
const PASSWORD = 'password';
const BASE_URL = 'http://maintenance-api:8080';


export function setup() {
  // register a new user and authenticate via a Bearer token.
  const res = http.post(`${BASE_URL}/users`, JSON.stringify({
    "nickname": NICKNAME,
    "email": EMAIL,
    "password": PASSWORD,
  }));

  check(res, { 'create user': (r) => r.status === 201 });

  const techLoginRes = http.post(`${BASE_URL}/login`, JSON.stringify({
    "email": EMAIL,
    "password": PASSWORD,
  }));

  const techAuthToken = techLoginRes.json();
  check(techAuthToken, { 'login successfully': () => techAuthToken !== '' });

  const managerLoginRes = http.post(`${BASE_URL}/login`, JSON.stringify({
    "email": "luther@gmail.com",
    "password": PASSWORD,
  }));

  const managerAuthToken = managerLoginRes.json();
  check(managerAuthToken, { 'login successfully': () => managerAuthToken !== '' });

  return {techAuthToken, managerAuthToken};
}

export default (tokens) => {
  const requestConfigWithTechAuthToken = (tag) => ({
    headers: {
      Authorization: `Bearer ${tokens.techAuthToken}`,
    },
    tags: Object.assign(
      {},
      {
        name: 'PrivateTasks',
      },
      tag
    ),
  });
  
  const requestConfigWithManagerAuthToken = (tag) => ({
    headers: {
      Authorization: `Bearer ${tokens.managerAuthToken}`,
    },
    tags: Object.assign(
      {},
      {
        name: 'PrivateTasks',
      },
      tag
    ),
  });

  group('Create and modify tasks', () => {
    let URL = `${BASE_URL}/tasks`;

    group('Create tasks', () => {
      const payload = JSON.stringify({
        summary: `${randomString(200)}`,
        date: '2011-10-05T14:48:00Z',
      });

      const res = http.post(URL, payload, requestConfigWithTechAuthToken({ name: 'Create' }));

      if (check(res, { 'tasks created correctly': (r) => r.status === 201 })) {
        URL = `${URL}/${res.json('id')}`;
      } else {
        console.log(`unable to create a Task ${res.status} ${res.body}`);
        return;
      }
    });

    group('Update task', () => {
      const new_summary = `${randomString(2500)}`
      const payload = JSON.stringify({
        summary: new_summary,
        date: '2012-10-05T14:48:00Z',
      })
      const res = http.put(URL, payload, requestConfigWithTechAuthToken({ name: 'Update' }));
      const isSuccessfulUpdate = check(res, {
        'updates worked': () => res.status === 200,
        'updated names were correct': () => res.json('summary') === new_summary,
      });

      if (!isSuccessfulUpdate) {
        console.log(`unable to update the task ${res.status} ${res.body}`);
        return;
      }
    });

    group('Get task', () => {
      const res = http.get(URL, requestConfigWithTechAuthToken({ name: 'Get' }));
      const isSuccessfulUpdate = check(res, {
        'get task worked': () => res.status === 200
      });

      if (!isSuccessfulUpdate) {
        console.log(`unable to get the task ${res.status} ${res.body}`);
        return;
      }
    });

    group('Delete task', () => {
      const delRes = http.del(URL, null, requestConfigWithManagerAuthToken({ name: 'Delete' }));

      const isSuccessfulDelete = check(delRes, {
        'task was deleted correctly': () => delRes.status === 204,
      });

      if (!isSuccessfulDelete) {
        console.log(`unable to delete the task ${delRes.status} ${delRes.body}`);
        return;
      }
    });
  });

  sleep(1);
};