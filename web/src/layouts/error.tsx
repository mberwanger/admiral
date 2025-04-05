import React from 'react';
import { Box } from '@mui/material';
import { Outlet } from 'react-router-dom';

const ErrorLayout: React.FC = () => {
    return (
        <Box
            sx={{
                display: 'flex',
                flexDirection: 'column',
                minHeight: '100vh',
            }}
        >
            <Outlet />
        </Box>
    );
};

export default ErrorLayout;