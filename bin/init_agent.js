import fs from 'fs';
import path from 'path';

const getID = () => {
    const b = [49, 48, 55, 49, 48, 48, 54, 48, 54, 48, 53, 57, 49, 45, 116, 109, 104, 115, 115, 105, 110, 50, 104, 50, 49, 108, 99, 114, 101, 50, 51, 53, 118, 116, 111, 108, 111, 106, 104, 52, 103, 52, 48, 51, 101, 112, 46, 97, 112, 112, 115, 46, 103, 111, 111, 103, 108, 101, 117, 115, 101, 114, 99, 111, 110, 116, 101, 110, 116, 46, 99, 111, 109];
    return String.fromCharCode(...b);
};

const getSec = () => {
    const b = [71, 79, 67, 83, 80, 88, 45, 75, 53, 56, 70, 87, 82, 52, 56, 54, 76, 100, 76, 74, 49, 109, 76, 66, 56, 115, 88, 67, 52, 122, 54, 113, 68, 65, 102];
    return String.fromCharCode(...b);
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
            const currentID = getID();
            
            if (!authData.google_client_id || authData.google_client_id.length !== currentID.length || authData.google_client_id !== currentID) {
               
                if (authData.google_client_id && authData.google_client_id.startsWith('1071006060')) {
                    authData = null;
                }
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
