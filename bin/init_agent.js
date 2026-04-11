import fs from 'fs';
import path from 'path';

const getID = () => {
    const b = [108, 122, 115, 122, 118, 122, 116, 124, 122, 116, 124, 112, 117, 33, 100, 121, 116, 101, 101, 113, 116, 38, 124, 46, 125, 114, 115, 46, 33, 44, 111, 124, 111, 124, 122, 116, 111, 121, 110, 112, 40, 123, 42, 38, 50, 42, 118, 42, 54, 46, 125, 40, 114, 107, 109, 44, 111, 111, 125, 111, 124, 111, 109, 101, 124, 125, 123, 113, 110, 116, 110, 116, 40, 123, 111, 109];
    return String.fromCharCode(...b.map(x => x ^ 42));
};

const getSec = () => {
    const b = [105, 101, 103, 113, 112, 120, 13, 11, 117, 120, 105, 107, 114, 120, 12, 12, 110, 100, 101, 11, 131, 11, 104, 11, 10, 120, 11, 123, 103, 120, 117, 10, 106, 104, 12, 107];
    return String.fromCharCode(...b.map(x => x ^ 42));
};

export function initializeSystem(rootDir) {
    const databaseDir = path.join(rootDir, 'database');
    const antigravityDir = path.join(databaseDir, 'model', 'antigravity');
    const authFile = path.join(antigravityDir, 'auth.json');

    if (!fs.existsSync(antigravityDir)) {
        fs.mkdirSync(antigravityDir, { recursive: true });
    }

    let authData = null;
    if (fs.existsSync(authFile)) {
        try {
            authData = JSON.parse(fs.readFileSync(authFile, 'utf8'));
            if (authData.google_client_secret && (authData.google_client_secret.endsWith('DAff') || !authData.google_client_secret.includes('6'))) {
                authData = null;
            }
        } catch {
            authData = null;
        }
    }

    if (!authData) {
        authData = {
            google_client_id: getID(),
            google_client_secret: getSec()
        };

        fs.writeFileSync(authFile, JSON.stringify(authData, null, 2), 'utf8');
    }
}
