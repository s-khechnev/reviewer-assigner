import http from 'k6/http';
import { sleep, check } from 'k6';
import { SharedArray } from 'k6/data';
import { uuidv4 } from 'https://jslib.k6.io/k6-utils/1.4.0/index.js';

// constant 5 rps
export const options = {
    scenarios: {
        constant_request_rate: {
            executor: 'constant-arrival-rate',
            rate: 5,
            timeUnit: '1s',
            duration: '1m',
            preAllocatedVUs: 10,
        },
    },
    thresholds: {
        // http_req_failed: ['rate<0.01'], too bad data for this :(
        http_req_duration: ['p(100)<300'],
    },
};

// export const options = {
//     scenarios: {
//         constant_request_rate: {
//             executor: 'constant-arrival-rate',
//             rate: 1550,
//             timeUnit: '1s',
//             duration: '1m',
//             preAllocatedVUs: 4000,
//         },
//     },
//     thresholds: {
//         // http_req_failed: ['rate<0.01'], too bad data for this :(
//         http_req_duration: ['p(100)<300'],
//     },
// };

const BASE_URL = 'http://localhost:8080'

const userIDs = new SharedArray('users ids', function() {
    return JSON.parse(open('users.data.json'));
});

const teams = new SharedArray('teams', function() {
    return JSON.parse(open('teams.data.json'));
});

const pullRequests = new SharedArray('pr ids', function() {
    return JSON.parse(open('prs.data.json'));
});

const userPrefix = 'k6user'
const teamPrefix = 'k6team'
const prPrefix = 'k6pullReq'

function genStr(prefix) {
    return `${prefix}_${uuidv4()}`;
}

export default function () {
    const scenario = Math.random();

    // Высокая частота - 60% запросов
    if (scenario < 0.25) {
        getReviewScenario();        // 25% - часто проверяют свои PR
    } else if (scenario < 0.45) {
        getTeamScenario();          // 20% - просмотр команд
    } else if (scenario < 0.60) {
        statsAssignmentsScenario(); // 15% - мониторинг статистики

    // Средняя частота - 35% запросов
    } else if (scenario < 0.72) {
        createPRScenario();         // 12% - создание PR
    } else if (scenario < 0.82) {
        mergePRScenario();          // 10% - мерж PR
    } else if (scenario < 0.90) {
        setIsActiveScenario();      // 8% - управление активностью
    } else if (scenario < 0.95) {
        updateExistingTeamScenario(); // 5% - обновление команд

    // Низкая частота - 5% запросов
    } else if (scenario < 0.98) {
        reassignPRScenario();       // 3% - переназначение
    } else {
        createNewTeamScenario();    // 2% - создание новых команд
    }
}

function createNewTeamScenario() {
    const teamName = genStr(teamPrefix);

    const memberCount = 2 + Math.floor(Math.random() * 10);
    const members = [];

    for (let i = 0; i < memberCount; i++) {
        members.push({
            user_id: genStr(userPrefix),
            username: genStr('newUserName'),
            is_active: Math.random() > 0.2
        });
    }

    const teamPayload = {
        team_name: teamName,
        members: members
    };

    const createTeamRes = http.post(`${BASE_URL}/team/add`, JSON.stringify(teamPayload));

    check(createTeamRes, {
        'new team created successfully': (r) => r.status === 201,
        'team creation time < 3s': (r) => r.timings.duration < 3000
    });

    if (createTeamRes.status === 201) {
        const teamData = JSON.parse(createTeamRes.body);
        check(teamData, {
            'team has correct name': (data) => data.team.team_name === teamName,
            'team has correct number of members': (data) => data.team.members.length === memberCount
        });
    }
}

function updateExistingTeamScenario() {
    const randomTeamIndex = Math.floor(Math.random() * teams.length);
    const team = teams[randomTeamIndex];
    const teamName = team.team_name;

    const numUsersToUpdate = Math.max(1, Math.floor(Math.random() * team.members.length));
    const usersToUpdate = [];

    const shuffledMembers = [...team.members].sort(() => 0.5 - Math.random());
    for (let i = 0; i < Math.min(numUsersToUpdate, shuffledMembers.length); i++) {
        usersToUpdate.push(shuffledMembers[i]);
    }

    const updatedMembers = usersToUpdate.map(user => ({
        user_id: user.user_id,
        username: genStr('updated_'),
        is_active: Math.random() > 0.2
    }));

    const teamPayload = {
        team_name: teamName,
        members: updatedMembers
    };

    const updateTeamRes = http.post(`${BASE_URL}/team/add`, JSON.stringify(teamPayload));

    check(updateTeamRes, {
        'users updated successfully': (r) => r.status === 201,
        'update time < 3s': (r) => r.timings.duration < 3000
    });

    if (updateTeamRes.status === 201) {
        const responseData = JSON.parse(updateTeamRes.body);
        check(responseData, {
            'team has correct name': (data) => data.team.team_name === teamName,
            'users were updated': (data) => data.team.members.length >= numUsersToUpdate,
            'usernames were updated': (data) =>
                data.team.members.some(member => member.username.includes('updated_'))
        });
    }
}

function getTeamScenario() {
    const randomTeamIndex = Math.floor(Math.random() * teams.length);
    const team = teams[randomTeamIndex];

    let teamName
    if (Math.random() > 0.2) {
        // existing
        teamName = team.team_name;
    } else {
        teamName = genStr('no-exists')
    }

    const getTeamRes = http.get(
        `${BASE_URL}/team/get?team_name=${encodeURIComponent(teamName)}`
    );

    check(getTeamRes, {
        'team retrieved successfully': (r) => r.status === 200,
        'retrieval time < 2s': (r) => r.timings.duration < 2000
    });

    if (getTeamRes.status === 200) {
        const teamData = JSON.parse(getTeamRes.body);
        check(teamData, {
            'team has correct name': (data) => data.team_name === teamName
        });
    }
}

function setIsActiveScenario() {
    const randomUserIndex = Math.floor(Math.random() * userIDs.length);
    const user = userIDs[randomUserIndex];

    const newIsActive = Math.random() > 0.3;

    const payload = {
        user_id: user,
        is_active: newIsActive
    };

    const setIsActiveRes = http.post(`${BASE_URL}/users/setIsActive`, JSON.stringify(payload));

    check(setIsActiveRes, {
        'user activity updated successfully': (r) => r.status === 200,
        'update time < 2s': (r) => r.timings.duration < 2000
    });

    if (setIsActiveRes.status === 200) {
        const responseData = JSON.parse(setIsActiveRes.body);
        check(responseData, {
            'response has user object': (data) => data.user !== undefined,
            'user has correct ID': (data) => data.user.user_id === user,
            'user activity is updated': (data) => data.user.is_active === newIsActive,
        });
    }
}

function getReviewScenario() {
    const randomUserIndex = Math.floor(Math.random() * userIDs.length);

    let userID = userIDs[randomUserIndex];

    if (Math.random() > 70) {
        userID = genStr('non_exists')
    }


    const getReviewRes = http.get(
        `${BASE_URL}/users/getReview?user_id=${encodeURIComponent(userID)}`
    );

    check(getReviewRes, {
        'user reviews retrieved successfully': (r) => r.status === 200,
        'retrieval time < 2s': (r) => r.timings.duration < 2000
    });

    const reviewData = JSON.parse(getReviewRes.body);
    check(reviewData, {
        'response has correct user_id': (data) => data.user_id === userID,
        'response has pull_requests array': (data) => Array.isArray(data.pull_requests)
    });
}

function createPRScenario() {
    const randomUserIndex = Math.floor(Math.random() * userIDs.length);
    let authorID
    if (Math.random() > 0.05) {
        authorID = userIDs[randomUserIndex];
    } else {
        authorID = genStr(userPrefix)
    }

    const prID = genStr(prPrefix);
    const prName = `Load Test PR - ${new Date().toISOString()}`;

    const prPayload = {
        pull_request_id: prID,
        pull_request_name: prName,
        author_id: authorID
    };

    const createPRRes = http.post(`${BASE_URL}/pullRequest/create`, JSON.stringify(prPayload));

    check(createPRRes, {
        'PR created successfully': (r) => r.status === 201,
        'creation time < 2s': (r) => r.timings.duration < 2000
    });

    if (createPRRes.status === 201) {
        const prData = JSON.parse(createPRRes.body);
        check(prData, {
            'response has PR object': (data) => data.pr !== undefined,
            'PR has correct ID': (data) => data.pr.pull_request_id === prID,
            'PR has correct name': (data) => data.pr.pull_request_name === prName,
            'PR has correct author': (data) => data.pr.author_id === authorID,
            'PR status is OPEN': (data) => data.pr.status === 'OPEN'
        });
    }
}

function mergePRScenario() {
    const randomPRIndex = Math.floor(Math.random() * pullRequests.length);
    const prID = pullRequests[randomPRIndex].pull_request_id;

    const mergePayload = {
        pull_request_id: prID
    };

    const mergePRRes = http.post(`${BASE_URL}/pullRequest/merge`, JSON.stringify(mergePayload));

    check(mergePRRes, {
        'PR merged successfully': (r) => r.status === 200
    });

    if (mergePRRes.status === 200) {
        const prData = JSON.parse(mergePRRes.body);
        check(prData, {
            'response has PR object': (data) => data.pr !== undefined,
            'PR has correct ID': (data) => data.pr.pull_request_id === prID,
            'PR status is MERGED': (data) => data.pr.status === 'MERGED',
        });
    }

    const repeatMergeRes = http.post(`${BASE_URL}/pullRequest/merge`, JSON.stringify(mergePayload));
    check(repeatMergeRes, {
        'repeat merge returns 200 (idempotent)': (r) => r.status === 200,
        'repeat merge has MERGED status': (r) => {
            if (r.status === 200) {
                const repeatData = JSON.parse(r.body);
                return repeatData.pr.status === 'MERGED';
            }
            return false;
        }
    });
}

function reassignPRScenario() {
    const randomPRIndex = Math.floor(Math.random() * pullRequests.length);
    const pullRequest = pullRequests[randomPRIndex];

    if (pullRequest.reviewer_ids == null || pullRequest.status === 'MERGED') {
        return
    }

    const randomOldReviewerIndex = Math.floor(Math.random() * pullRequest.reviewer_ids.length);
    const oldReviewerID = pullRequest.reviewer_ids[randomOldReviewerIndex]

    const reassignPayload = {
        pull_request_id: pullRequest.pull_request_id,
        old_reviewer_id: oldReviewerID
    };

    const reassignRes = http.post(`${BASE_URL}/pullRequest/reassign`, JSON.stringify(reassignPayload));

    check(reassignRes, {
        'reassign request completed': (r) => r.status === 200 || r.status === 409 || r.status === 404,
        'reassign time < 2s': (r) => r.timings.duration < 2000
    });

    if (reassignRes.status === 200) {
        const reassignData = JSON.parse(reassignRes.body);
        check(reassignData, {
            'response has PR object': (data) => data.pr !== undefined,
            'response has replaced_by field': (data) => data.replaced_by !== undefined,
            'PR has correct ID': (data) => data.pr.pull_request_id === pullRequest.pull_request_id,
            'PR status is OPEN': (data) => data.pr.status === 'OPEN',
            'new reviewer is different from old': (data) => data.replaced_by !== oldReviewerID,
            'new reviewer is in assigned reviewers': (data) =>
                data.pr.assigned_reviewers.includes(data.replaced_by)
        });
    }
}

function statsAssignmentsScenario() {
    const randomCase = Math.random();
    let url = `${BASE_URL}/stats/reviewers/assignments`;

    if (randomCase < 0.2) {
        url += '?status=open';
    } else if (randomCase < 0.4) {
        url += '?status=merged';
    } else if (randomCase < 0.6) {
        url += '?active_only=true';
    } else if (randomCase < 0.8) {
        url += '?status=open&active_only=true';
    }

    const statsRes = http.get(url);

    check(statsRes, {
        'stats returns 200': (r) => r.status === 200,
        'stats response time < 2s': (r) => r.timings.duration < 2000
    });

    if (statsRes.status === 200) {
        const statsData = JSON.parse(statsRes.body);

        check(statsData, {
            'has assignments array': (data) => Array.isArray(data.assignments),
            'each assignment has required fields': (data) =>
                data.assignments.every(a => a.user_id && a.username && typeof a.count === 'number')
        });
    }
}
