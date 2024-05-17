
import { useUserInfo } from '@/hooks/users';
import { router } from 'expo-router';
import { useEffect } from 'react';
import { Snackbar } from 'react-native-paper';

export default function Index() {

    const { userInfo, error, loading } = useUserInfo();


    useEffect(() => {
        if (loading) return

        if (userInfo) {
            router.replace('/dashboard')
        } else {
            router.replace('/signin')
        }

    }, [userInfo, loading])

    {
        error && <Snackbar
            visible
            onDismiss={() => router.replace('/signin')}
            action={{
                label: 'Hide',
            }}>
            {error.message}
        </Snackbar>
    }


    return null
}

