import { UserInfo, getUserInfo } from '@/api/users';
import { useEffect, useState } from 'react';


export const useUserInfo = () => {
    const [userInfo, setUserInfo] = useState<UserInfo|null>(null);
    const [loading, setLoading] = useState(true);
    const [error, setError] = useState<any>(null);

    useEffect(() => {
        const fetchUserInfo = async () => {
            try {
                const user = await getUserInfo();
                setUserInfo(user);
            } catch (err) {
                setError(err);
            } finally {
                setLoading(false);
            }
        };

        fetchUserInfo();
    }, []);

    return { userInfo, loading, error };
};

